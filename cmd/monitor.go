package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type PrometheusResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Values [][]interface{}   `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

type ResourceRecommendation struct {
	Namespace         string
	PodName           string
	ContainerName     string
	CurrentCPU        string
	CurrentMemory     string
	RecommendedCPU    string
	RecommendedMemory string
	Reason            string
}

var (
	prometheusURL string
	watchInterval int
	cpuThreshold  float64
	memThreshold  float64
	autoApply     bool
)

func init() {
	rootCmd.AddCommand(monitorCmd)

	monitorCmd.Flags().StringVar(&prometheusURL, "prometheus-url", "http://localhost:9090", "Prometheus server URL")
	monitorCmd.Flags().IntVar(&watchInterval, "interval", 30, "Monitoring interval in seconds")
	monitorCmd.Flags().Float64Var(&cpuThreshold, "cpu-threshold", 0.8, "CPU utilization threshold (0.0-1.0)")
	monitorCmd.Flags().Float64Var(&memThreshold, "memory-threshold", 0.8, "Memory utilization threshold (0.0-1.0)")
	monitorCmd.Flags().BoolVar(&autoApply, "auto-apply", false, "Automatically apply recommendations")
}

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Monitor resources and provide resize recommendations",
	Long:  "Continuously monitor pod resources using Prometheus metrics and suggest optimizations",
	RunE: func(cmd *cobra.Command, args []string) error {
		return startMonitoring()
	},
}

func startMonitoring() error {
	clientset, err := getClientSet("")
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	fmt.Printf("🔍 Starting resource monitoring...\n")
	fmt.Printf("📊 Prometheus URL: %s\n", prometheusURL)
	fmt.Printf("⏰ Check interval: %d seconds\n", watchInterval)
	fmt.Printf("🎯 CPU threshold: %.1f%%\n", cpuThreshold*100)
	fmt.Printf("💾 Memory threshold: %.1f%%\n", memThreshold*100)
	fmt.Printf("🤖 Auto-apply: %v\n\n", autoApply)

	ticker := time.NewTicker(time.Duration(watchInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := analyzeResources(clientset)
			if err != nil {
				log.Printf("Error analyzing resources: %v", err)
			}
		}
	}
}

func analyzeResources(clientset *kubernetes.Clientset) error {
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list pods: %v", err)
	}

	recommendations := []ResourceRecommendation{}

	for _, pod := range pods.Items {
		if pod.Status.Phase != v1.PodRunning {
			continue
		}

		for _, container := range pod.Spec.Containers {
			cpuUsage, memUsage, err := getContainerMetrics(pod.Namespace, pod.Name, container.Name)
			if err != nil {
				log.Printf("Failed to get metrics for %s/%s/%s: %v", pod.Namespace, pod.Name, container.Name, err)
				continue
			}

			recommendation := analyzeContainerUsage(pod, container, cpuUsage, memUsage)
			if recommendation != nil {
				recommendations = append(recommendations, *recommendation)
			}
		}
	}

	if len(recommendations) > 0 {
		displayRecommendations(recommendations)
		if autoApply {
			applyRecommendations(clientset, recommendations)
		}
	} else {
		fmt.Printf("✅ All resources are optimally sized (checked %d pods)\n", len(pods.Items))
	}

	return nil
}

func getContainerMetrics(namespace, podName, containerName string) (float64, float64, error) {
	// Query CPU usage
	cpuQuery := fmt.Sprintf(`rate(container_cpu_usage_seconds_total{namespace="%s",pod="%s",container="%s"}[5m])`,
		namespace, podName, containerName)
	cpuUsage, err := queryPrometheus(cpuQuery)
	if err != nil {
		return 0, 0, err
	}

	// Query Memory usage
	memQuery := fmt.Sprintf(`container_memory_usage_bytes{namespace="%s",pod="%s",container="%s"}`,
		namespace, podName, containerName)
	memUsage, err := queryPrometheus(memQuery)
	if err != nil {
		return cpuUsage, 0, err
	}

	return cpuUsage, memUsage, nil
}

func queryPrometheus(query string) (float64, error) {
	encodedQuery := url.QueryEscape(query)
	url := fmt.Sprintf("%s/api/v1/query?query=%s", prometheusURL, encodedQuery)

	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var promResp PrometheusResponse
	if err := json.NewDecoder(resp.Body).Decode(&promResp); err != nil {
		return 0, err
	}

	if len(promResp.Data.Result) == 0 {
		return 0, nil
	}

	valueStr, ok := promResp.Data.Result[0].Values[len(promResp.Data.Result[0].Values)-1][1].(string)
	if !ok {
		return 0, fmt.Errorf("invalid metric value format")
	}

	return strconv.ParseFloat(valueStr, 64)
}

func analyzeContainerUsage(pod v1.Pod, container v1.Container, cpuUsage, memUsage float64) *ResourceRecommendation {
	currentCPU := container.Resources.Requests.Cpu()
	currentMem := container.Resources.Requests.Memory()

	if currentCPU == nil || currentMem == nil {
		return nil
	}

	cpuRequestMillis := float64(currentCPU.MilliValue())
	memRequestBytes := float64(currentMem.Value())

	var needsResize bool
	var reason string
	var recommendedCPU, recommendedMem string

	// Analyze CPU
	cpuUtil := cpuUsage / (cpuRequestMillis / 1000)
	if cpuUtil > cpuThreshold {
		needsResize = true
		newCPUMillis := int64(cpuUsage * 1000 * 1.2) // 20% buffer
		recommendedCPU = fmt.Sprintf("%dm", newCPUMillis)
		reason += fmt.Sprintf("CPU utilization: %.1f%% ", cpuUtil*100)
	} else if cpuUtil < 0.2 { // Under-utilized
		needsResize = true
		newCPUMillis := int64(cpuUsage * 1000 * 1.5) // 50% buffer for safety
		if newCPUMillis < 100 {
			newCPUMillis = 100 // Minimum 100m
		}
		recommendedCPU = fmt.Sprintf("%dm", newCPUMillis)
		reason += fmt.Sprintf("CPU under-utilized: %.1f%% ", cpuUtil*100)
	} else {
		recommendedCPU = currentCPU.String()
	}

	// Analyze Memory
	memUtil := memUsage / memRequestBytes
	if memUtil > memThreshold {
		needsResize = true
		newMemMB := int64(memUsage * 1.2 / 1024 / 1024) // 20% buffer
		recommendedMem = fmt.Sprintf("%dMi", newMemMB)
		reason += fmt.Sprintf("Memory utilization: %.1f%% ", memUtil*100)
	} else if memUtil < 0.3 { // Under-utilized
		needsResize = true
		newMemMB := int64(memUsage * 1.5 / 1024 / 1024) // 50% buffer
		if newMemMB < 128 {
			newMemMB = 128 // Minimum 128Mi
		}
		recommendedMem = fmt.Sprintf("%dMi", newMemMB)
		reason += fmt.Sprintf("Memory under-utilized: %.1f%% ", memUtil*100)
	} else {
		recommendedMem = currentMem.String()
	}

	if !needsResize {
		return nil
	}

	return &ResourceRecommendation{
		Namespace:         pod.Namespace,
		PodName:           pod.Name,
		ContainerName:     container.Name,
		CurrentCPU:        currentCPU.String(),
		CurrentMemory:     currentMem.String(),
		RecommendedCPU:    recommendedCPU,
		RecommendedMemory: recommendedMem,
		Reason:            strings.TrimSpace(reason),
	}
}

func displayRecommendations(recommendations []ResourceRecommendation) {
	fmt.Printf("\n🎯 Resource Resize Recommendations (%s)\n", time.Now().Format("15:04:05"))
	fmt.Printf("═══════════════════════════════════════════════════════════════\n")

	for _, rec := range recommendations {
		fmt.Printf("\n📦 Pod: %s/%s (container: %s)\n", rec.Namespace, rec.PodName, rec.ContainerName)
		fmt.Printf("   Current:     CPU: %-8s Memory: %-8s\n", rec.CurrentCPU, rec.CurrentMemory)
		fmt.Printf("   Recommended: CPU: %-8s Memory: %-8s\n", rec.RecommendedCPU, rec.RecommendedMemory)
		fmt.Printf("   Reason: %s\n", rec.Reason)
	}
	fmt.Println()
}

func applyRecommendations(clientset *kubernetes.Clientset, recommendations []ResourceRecommendation) {
	fmt.Printf("🤖 Auto-applying %d recommendations...\n", len(recommendations))
	// Note: In a real scenario, you'd update deployments/statefulsets, not pods directly
	// This would require finding the parent controller and updating its spec
	fmt.Printf("⚠️  Auto-apply is simulated - would update parent controllers in production\n")
}
