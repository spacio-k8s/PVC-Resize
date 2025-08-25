package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "resize-assistant",
	Short: "Kubernetes resource resize assistant for Minikube",
	Long: `A CLI tool that monitors Kubernetes resources in Minikube and provides 
intelligent resizing suggestions based on Prometheus metrics.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Use 'resize-assistant --help' to see available commands")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing resize-assistant: %v\n", err)
		os.Exit(1)
	}
}
