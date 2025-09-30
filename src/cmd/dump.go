package cmd

import (
	"fmt"
	internal "pvc-audit/Internal"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dump dummy data to PVC, tracks usage information (used size, allocated size, wastage)",
	RunE: func(cmd *cobra.Command, args []string) error {
		pvcName, _ := cmd.Flags().GetString("pvc")
		if pvcName == "" {
			return fmt.Errorf("PVC name is required")
		}

		sizeFlag, _ := cmd.Flags().GetString("size") // optional test data size

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

		for _, ns := range namespaces {
			pvcs, err := internal.ListPVCs(ns)
			if err != nil {
				fmt.Printf("Error listing PVCs in namespace '%s': %v\n", ns, err)
				continue
			}

			for _, pvc := range pvcs {
				if pvc.Name != pvcName {
					continue
				}

				allocated := pvc.Status.Capacity.Storage().Value() / (1024 * 1024) // MB

				// find pod and mount path
				podName, mountPath, err := internal.FindPodAndMountPathForPVC(ns, pvcName)
				if err != nil {
					fmt.Printf("PVC '%s' in namespace '%s': allocated=%d MB, no pod using it\n", pvcName, ns, allocated)
					continue
				}

				// write test data if --size is specified
				if sizeFlag != "" {
					sizeMB := parseSizeToMB(sizeFlag)
					fmt.Printf("⏳ Writing %d MB of test data to PVC '%s' in pod '%s'\n", sizeMB, pvcName, podName)

					writeCmd := fmt.Sprintf("sh -c 'dd if=/dev/zero of=%s/testfile bs=1M count=%d conv=fsync'", mountPath, sizeMB)
					out, err := internal.ExecInPod(podName, ns, writeCmd)
					if err != nil {
						fmt.Printf("Error writing test data: %v\nOutput:\n%s\n", err, out)
						continue
					}
					fmt.Printf("Output of dd command:\n%s\n", out)
				}

				// get used size
				usedStr, err := internal.ExecInPod(podName, ns, fmt.Sprintf("du -sm %s | cut -f1", mountPath))
				if err != nil {
					fmt.Printf("Error getting used size for PVC '%s' in pod '%s': %v\n", pvcName, podName, err)
					continue
				}

				usedStr = strings.TrimSpace(usedStr)
				usedInt, err := strconv.Atoi(usedStr)
				if err != nil {
					fmt.Printf("Error converting used size to int: %v (raw='%s')\n", err, usedStr)
					continue
				}

				used := int64(usedInt)
				wasted := allocated - used
				wastePct := (wasted * 100) / allocated

				fmt.Printf("PVC '%s' in namespace '%s':\n", pvcName, ns)
				fmt.Printf("  Allocated Size : %d MB\n", allocated)
				fmt.Printf("  Used Size      : %d MB\n", used)
				fmt.Printf("  Wasted Space   : %d MB (%d%%)\n", wasted, wastePct)
				fmt.Printf("  Mounted Pod    : %s at %s\n", podName, mountPath)
			}
		}

		return nil
	},
}

func parseSizeToMB(size string) int {
	size = strings.TrimSpace(strings.ToUpper(size))
	if strings.HasSuffix(size, "MB") {
		val, _ := strconv.Atoi(strings.TrimSuffix(size, "MB"))
		return val
	} else if strings.HasSuffix(size, "M") {
		val, _ := strconv.Atoi(strings.TrimSuffix(size, "M"))
		return val
	}
	val, _ := strconv.Atoi(size)
	return val
}

func init() {
	rootCmd.AddCommand(dumpCmd)
	dumpCmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Kubernetes namespace")
	dumpCmd.Flags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "Dump PVC info in all namespaces")
	dumpCmd.Flags().StringP("pvc", "p", "", "PVC name")
	dumpCmd.Flags().StringP("size", "s", "", "Optional: fill PVC with test data (MB)")
	dumpCmd.MarkFlagRequired("pvc")
}
