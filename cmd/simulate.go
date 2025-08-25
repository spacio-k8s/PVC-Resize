package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	simNamespace string
	simWorkloads int
	simPVCs      int
	cleanupSim   bool
)

func init() {
	rootCmd.AddCommand(simulateCmd)

	simulateCmd.Flags().StringVarP(&simNamespace, "namespace", "n", "resize-test", "Namespace for simulation")
	simulateCmd.Flags().IntVar(&simWorkloads, "workloads", 3, "Number of workloads to create")
	simulateCmd.Flags().IntVar(&simPVCs, "pvcs", 2, "Number of PVCs to create")
	simulateCmd.Flags().BoolVar(&cleanupSim, "cleanup", false, "Clean up simulation resources")
}

var simulateCmd = &cobra.Command{
	Use:   "simulate",
	Short: "Create simulation workloads and PVCs in Minikube",
	Long:  "Deploy sample applications with various resource patterns for testing resize recommendations",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cleanupSim {
			return cleanupSimulation()
		}
		return createSimulation()
	},
}

func createSimulation() error {
	clientset, err := getClientSet("")
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	fmt.Printf("🚀 Creating simulation in namespace '%s'...\n", simNamespace)

	// Create namespace
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: simNamespace,
		},
	}
	_, err = clientset.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})
	if err != nil {
		fmt.Printf("⚠️  Namespace might already exist: %v\n", err)
	}

	// Create PVCs
	err = createSimulationPVCs(clientset)
	if err != nil {
		return err
	}

	// Create workloads
	err = createSimulationWorkloads(clientset)
	if err != nil {
		return err
	}

	fmt.Printf("✅ Simulation created successfully!\n")
	fmt.Printf("💡 Use 'resize-assistant monitor' to start monitoring\n")
	return nil
}

func createSimulationPVCs(clientset *kubernetes.Clientset) error {
	fmt.Printf("💾 Creating %d PVCs...\n", simPVCs)

	for i := 0; i < simPVCs; i++ {
		// Create PVCs with different usage patterns
		var size, usage string
		var overprovisioned bool

		switch i % 3 {
		case 0: // Overprovisioned
			size = "10Gi"
			usage = "2.5Gi"
			overprovisioned = true
		case 1: // Well-sized
			size = "5Gi"
			usage = "4.2Gi"
		case 2: // Underprovisioned
			size = "3Gi"
			usage = "2.8Gi"
		}

		pvc := &v1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("sim-pvc-%d", i),
				Namespace: simNamespace,
				Annotations: map[string]string{
					"resize-cltool/used-storage": usage,
				},
			},
			Spec: v1.PersistentVolumeClaimSpec{
				AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceStorage: resource.MustParse(size),
					},
				},
			},
		}

		_, err := clientset.CoreV1().PersistentVolumeClaims(simNamespace).Create(context.TODO(), pvc, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create PVC: %v", err)
		}

		status := "📏"
		if overprovisioned {
			status = "📈"
		}
		fmt.Printf("  %s Created PVC sim-pvc-%d: %s (used: %s)\n", status, i, size, usage)
	}
	return nil
}

func createSimulationWorkloads(clientset *kubernetes.Clientset) error {
	fmt.Printf("⚙️  Creating %d workloads...\n", simWorkloads)

	workloadTypes := []struct {
		name     string
		cpuReq   string
		memReq   string
		pattern  string
		replicas int32
	}{
		{"cpu-intensive", "500m", "256Mi", "High CPU usage", 1},
		{"memory-hungry", "100m", "1Gi", "High memory usage", 1},
		{"underutilized", "1000m", "2Gi", "Low resource usage", 2},
		{"balanced", "200m", "512Mi", "Balanced usage", 1},
		{"bursty", "300m", "768Mi", "Variable load", 1},
	}

	for i := 0; i < simWorkloads && i < len(workloadTypes); i++ {
		wl := workloadTypes[i]

		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("sim-%s", wl.name),
				Namespace: simNamespace,
				Labels: map[string]string{
					"app":        fmt.Sprintf("sim-%s", wl.name),
					"simulation": "resize-assistant",
				},
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &wl.replicas,
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": fmt.Sprintf("sim-%s", wl.name),
					},
				},
				Template: v1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": fmt.Sprintf("sim-%s", wl.name),
						},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name:    "app",
								Image:   "busybox:latest",
								Command: []string{"sh", "-c"},
								Args:    []string{getWorkloadScript(wl.name)},
								Resources: v1.ResourceRequirements{
									Requests: v1.ResourceList{
										v1.ResourceCPU:    resource.MustParse(wl.cpuReq),
										v1.ResourceMemory: resource.MustParse(wl.memReq),
									},
									Limits: v1.ResourceList{
										v1.ResourceCPU:    resource.MustParse(wl.cpuReq),
										v1.ResourceMemory: resource.MustParse(wl.memReq),
									},
								},
							},
						},
					},
				},
			},
		}

		_, err := clientset.AppsV1().Deployments(simNamespace).Create(context.TODO(), deployment, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create deployment %s: %v", wl.name, err)
		}

		fmt.Printf("  🏗️  Created deployment sim-%s: CPU: %s, Memory: %s (%s)\n",
			wl.name, wl.cpuReq, wl.memReq, wl.pattern)
	}

	return nil
}

func getWorkloadScript(workloadType string) string {
	scripts := map[string]string{
		"cpu-intensive": "while true; do echo 'CPU intensive task'; for i in $(seq 1 10000); do echo $i > /dev/null; done; sleep 1; done",
		"memory-hungry": "while true; do echo 'Memory allocation'; dd if=/dev/zero of=/tmp/memtest bs=1M count=100 2>/dev/null; sleep 5; rm -f /tmp/memtest; done",
		"underutilized": "while true; do echo 'Low usage task'; sleep 30; done",
		"balanced":      "while true; do echo 'Balanced workload'; sleep 5; for i in $(seq 1 1000); do echo $i > /dev/null; done; done",
		"bursty":        "while true; do echo 'Bursty load'; sleep $((RANDOM % 20 + 5)); for i in $(seq 1 5000); do echo $i > /dev/null; done; done",
	}

	if script, exists := scripts[workloadType]; exists {
		return script
	}
	return "while true; do echo 'Default workload'; sleep 10; done"
}

func cleanupSimulation() error {
	clientset, err := getClientSet("")
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	fmt.Printf("🧹 Cleaning up simulation in namespace '%s'...\n", simNamespace)

	// Delete namespace (this will delete all resources within it)
	err = clientset.CoreV1().Namespaces().Delete(context.TODO(), simNamespace, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete namespace: %v", err)
	}

	fmt.Printf("✅ Simulation cleanup completed!\n")
	return nil
}
