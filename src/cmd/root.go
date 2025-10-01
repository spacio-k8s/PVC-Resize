package cmd

import "github.com/spf13/cobra"

// define at package level so all commands can see it
var (
	namespace     string
	allNamespaces bool
	rootCmd       = &cobra.Command{
		Use:   "spacio",
		Short: "Spacio PVC Auditor - Audit wasted PVC storage in Kubernetes clusters",
		Long: `Spacio PVC Auditor helps you identify:
  - Wasted PVC storage
  - Orphaned volumes
  - Over-provisioned PVCs
across your Kubernetes cluster.`,
	}
)

func Execute() error {
	return rootCmd.Execute()

}
