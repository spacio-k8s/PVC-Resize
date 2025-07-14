package cmd

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const usageAnnotation = "resize-cltool/used-storage"

var (
	threshold  float64
	kubeconfig string
	namespace  string
)

func init() {
	pvcCmd.AddCommand(overprovisionedCmd)
	rootCmd.AddCommand(pvcCmd)

	// Flags for the overprovisioned subcommand
	overprovisionedCmd.Flags().Float64VarP(&threshold, "threshold", "t", 2.0, "Overprovision threshold (e.g. 2.0 = 2x)")
	overprovisionedCmd.Flags().StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig (optional)")
	overprovisionedCmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Namespace to filter PVCs (default: all)")
}

var pvcCmd = &cobra.Command{
	Use:   "pvc",
	Short: "Trigger PVC operations (expand/shrink/dry-run/list-wastage)",
}

var overprovisionedCmd = &cobra.Command{
	Use:   "overprovisioned",
	Short: "List PVCs where requested storage is significantly more than used",
	RunE: func(cmd *cobra.Command, args []string) error {
		return findOverprovisionedPVCs()
	},
}

func findOverprovisionedPVCs() error {
	clientset, err := getClientSet(kubeconfig)
	if err != nil {
		return fmt.Errorf("❌ Failed to create Kubernetes client: %v", err)
	}

	pvcs, err := clientset.CoreV1().PersistentVolumeClaims(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("❌ Failed to list PVCs: %v", err)
	}

	for _, pvc := range pvcs.Items {
		processPVC(pvc, threshold)
	}

	return nil
}

func getClientSet(kubeconfigPath string) (*kubernetes.Clientset, error) {
	if kubeconfigPath != "" {
		flag.Set("logtostderr", "true")
		config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			return nil, err
		}
		return kubernetes.NewForConfig(config)
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func processPVC(pvc v1.PersistentVolumeClaim, threshold float64) {
	ns := pvc.Namespace
	name := pvc.Name

	requested := pvc.Spec.Resources.Requests.Storage()
	requestedGi := float64(requested.ScaledValue(resource.Giga))

	usedStr, ok := pvc.Annotations[usageAnnotation]
	if !ok {
		return // Skip if annotation is not present
	}

	usedGi, err := parseStorageToGi(usedStr)
	if err != nil {
		log.Printf("⚠️  Could not parse usage for %s/%s: %v\n", ns, name, err)
		return
	}

	if requestedGi > usedGi*threshold {
		fmt.Printf("🔍 PVC %s/%s is overprovisioned — Requested: %.2fGi, Used: %.2fGi\n", ns, name, requestedGi, usedGi)
	} else {
		fmt.Printf("✅ PVC %s/%s is within limits — Requested: %.2fGi, Used: %.2fGi\n", ns, name, requestedGi, usedGi)
	}
}

func parseStorageToGi(value string) (float64, error) {
	value = strings.TrimSpace(value)
	if strings.HasSuffix(value, "Gi") {
		v := strings.TrimSuffix(value, "Gi")
		return strconv.ParseFloat(v, 64)
	} else if strings.HasSuffix(value, "Mi") {
		v := strings.TrimSuffix(value, "Mi")
		mi, err := strconv.ParseFloat(v, 64)
		return mi / 1024, err
	}
	return 0, fmt.Errorf("unsupported storage unit: %s", value)
}
