package cmd

import (
	"fmt"
	"os"

	internal "pvc-audit/Internal"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

var podsCmd = &cobra.Command{
	Use:   "pods",
	Short: "List pods attached to PVCs (or show unattached PVCs)",
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

		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"Namespace", "PVC", "Pod(s)", "Attachment"})

		for _, ns := range namespaces {
			pvcs, err := internal.ListPVCs(ns)
			if err != nil {
				fmt.Printf("Error listing PVCs in namespace %s: %v\n", ns, err)
				continue
			}

			for _, pvc := range pvcs {
				pods, err := internal.FindPodsForPVC(ns, pvc.Name)
				if err != nil {
					fmt.Printf("Error finding pods for PVC %s/%s: %v\n", ns, pvc.Name, err)
					continue
				}

				if len(pods) == 0 {
					// unattached PVC
					t.AppendRow(table.Row{ns, pvc.Name, "-", "Unattached"})
				} else {
					// attached PVC
					t.AppendRow(table.Row{ns, pvc.Name, pods[0], "Attached"})
					for _, pod := range pods[1:] {
						// additional pods in separate rows
						t.AppendRow(table.Row{ns, pvc.Name, pod, "Attached"})
					}
				}
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
	rootCmd.AddCommand(podsCmd)
	podsCmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Kubernetes namespace")
	podsCmd.Flags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "List PVCs in all namespaces")
}
