package cmd

import "github.com/spf13/cobra"

// define at package level so all commands can see it
var (
	namespace     string
	allNamespaces bool
	rootCmd       = &cobra.Command{
		Use:   "pvc-audit",
		Short: "PVC Audit CLI",
	}
)

func Execute() error {
	return rootCmd.Execute()

}
