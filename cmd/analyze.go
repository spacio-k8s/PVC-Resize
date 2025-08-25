package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	rootCmd.AddCommand(analyzeCmd)
	analyzeCmd.AddCommand(clusterCmd)
	analyzeCmd.AddCommand(nodeCmd)
}

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze cluster and node resources",
	Long:  "Perform detailed analysis of cluster capacity and node utilization",
}

var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Analyze overall cluster resource utilization",
	RunE: func(cmd *cobra.Command, args []string) error {
		return analyzeCluster()
	},
}

var nodeCmd = &cobra.Command{
	Use:   "nodes",
	Short: "Analyze individual node resource utilization",
	RunE: func(cmd *cobra.Command, args []string) error {
		return analyzeNodes()
	},
}

func analyzeCluster() error {
	clientset, err := getClientSet("")
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	fmt.Printf("📊 Cluster Analysis\n")
	fmt.Printf("═══════════════════════════════════════\n")

	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list nodes: %v", err)
	}

	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list pods: %v", err)
	}

	storageClasses, err := clientset.StorageV1().StorageClasses().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list storage classes: %v", err)
	}

	fmt.Printf("🏗️  Total Nodes: %d\n", len(nodes.Items))
	fmt.Printf("📦 Total Pods: %d\n", len(pods.Items))
	fmt.Printf("💾 Storage Classes: ")
	for i, sc := range storageClasses.Items {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Print(sc.Name)
	}
	fmt.Println()
	fmt.Printf("🔄 Recommendations: Integration with Prometheus can enhance this further\n")

	return nil
}

func analyzeNodes() error {
	clientset, err := getClientSet("")
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	fmt.Printf("🖥️  Node Analysis\n")
	fmt.Printf("═══════════════════════════════════════\n")

	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list nodes: %v", err)
	}

	for _, node := range nodes.Items {
		cpu := node.Status.Allocatable["cpu"]
		mem := node.Status.Allocatable["memory"]
		storage := node.Status.Allocatable["ephemeral-storage"]

		fmt.Printf("📍 %s:\n", node.Name)
		fmt.Printf("  CPU: %s allocatable\n", cpu.String())
		fmt.Printf("  Memory: %s allocatable\n", mem.String())
		fmt.Printf("  Storage: %s allocatable\n", storage.String())
	}

	return nil
}
