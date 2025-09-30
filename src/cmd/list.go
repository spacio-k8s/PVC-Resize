package cmd

import (
	"fmt"
	"os"

	internal "pvc-audit/Internal"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List PVCs in a namespace or all namespaces",
	RunE: func(cmd *cobra.Command, args []string) error {
		namespaces := []string{}
		if allNamespaces {
			nsList, err := internal.ListNamespaces()
			if err != nil {
				return err
			}
			namespaces = nsList
		} else {
			namespaces = []string{namespace}
		}

		// Create one table across all namespaces
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"Namespace", "Name", "Allocated Storage"})

		for _, ns := range namespaces {
			pvcs, err := internal.ListPVCs(ns)
			if err != nil {
				fmt.Printf("Error listing PVCs in namespace %s: %v\n", ns, err)
				continue
			}

			if len(pvcs) == 0 {
				continue
			}

			for _, pvc := range pvcs {
				t.AppendRow(table.Row{
					ns,
					pvc.Name,
					pvc.Status.Capacity.Storage().String(),
				})
			}
		}

		if t.Length() == 0 {
			fmt.Println("No PVCs found")
			return nil
		}

		t.Render()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Kubernetes namespace")
	listCmd.Flags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "List PVCs in all namespaces")
}
