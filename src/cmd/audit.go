package cmd

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	Internal "pvc-audit/Internal"
	"pvc-audit/util"

	"github.com/spf13/cobra"
)

// ANSI colors for categories
func ColorizeCategory(cat string) string {
	switch cat {
	case "Critical":
		return "\033[41;37m Critical \033[0m" // red bg, white text
	case "Over-provisioned":
		return "\033[43;30m Overprovisioned \033[0m" // yellow bg, black text
	case "Unused":
		return "\033[44;37m Unused \033[0m" // blue bg, white text
	case "Healthy":
		return "\033[42;30m Healthy \033[0m" // green bg, black text
	default:
		return cat
	}
}

func GenerateCLIAuditReport(clusterReport ClusterReport) string {
	report := strings.Builder{}

	report.WriteString("\n📊 PVC Audit Summary Report\n")
	report.WriteString("─────────────────────────────────────────────\n")
	report.WriteString(fmt.Sprintf("Cluster Name             : %s\n", clusterReport.ClusterName))
	report.WriteString(fmt.Sprintf("Generated At             : %s\n", clusterReport.GeneratedAt))
	report.WriteString(fmt.Sprintf("Total Namespaces Audited : %d\n", clusterReport.TotalNamespaces))
	report.WriteString(fmt.Sprintf("Total PVCs Audited       : %d\n\n", clusterReport.TotalPVCs))

	report.WriteString("🧱 PVC Space Summary\n")
	report.WriteString("─────────────────────────────────────────────\n")

	unit := "GB"
	allocatedVal := clusterReport.TotalAllocatedGB
	usedVal := clusterReport.TotalUsedGB
	wastedVal := clusterReport.TotalWastedGB

	// Fallback to MB if allocated < 1 GB
	if clusterReport.TotalAllocatedGB < 1 {
		unit = "MB"
		allocatedVal = clusterReport.TotalAllocatedGB * 1024
		usedVal = clusterReport.TotalUsedGB * 1024
		wastedVal = clusterReport.TotalWastedGB * 1024
	}

	report.WriteString(fmt.Sprintf("Total Allocated Space (%s) : %.2f %s\n", unit, allocatedVal, unit))
	report.WriteString(fmt.Sprintf("Total Used Space (%s)      : %.2f %s\n", unit, usedVal, unit))
	report.WriteString(fmt.Sprintf("Total Wasted Space (%s)    : %.2f %s\n", unit, wastedVal, unit))

	wastagePct := 0.0
	if clusterReport.TotalAllocatedGB > 0 {
		wastagePct = (clusterReport.TotalWastedGB / clusterReport.TotalAllocatedGB) * 100
	}
	report.WriteString(fmt.Sprintf("Wastage Percentage          : %.1f%%\n\n", wastagePct))

	report.WriteString("⚠️ PVC Wastage Details\n")
	report.WriteString("─────────────────────────────────────────────\n")
	report.WriteString(fmt.Sprintf("PVCs with High Wastage (≥80%%) : %d\n", clusterReport.PVCsWithWastage))
	report.WriteString(fmt.Sprintf("Unattached PVCs                : %d\n", len(clusterReport.UnattachedPVCs)))
	report.WriteString(fmt.Sprintf("Cleanup Candidates             : %d\n\n", len(clusterReport.CleanupCandidates)))

	report.WriteString("📋 Top 5 High Wastage PVCs\n")
	report.WriteString("──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────\n")
	report.WriteString(fmt.Sprintf("| %-15s | %-33s | %-10s | %-9s | %-9s | %-9s | %-12s | %-15s |\n",
		"Namespace", "PVC Name", "Allocated", "Used", "Wasted", "Used (%)", "Wastage (%)", "Category"))
	report.WriteString("──────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────\n")
	count := 0
	for _, nsReport := range clusterReport.NamespaceReports {
		for _, pvc := range nsReport.PVCs {
			if pvc.WastagePct >= 80 {
				allocVal, usedVal, wastedVal, unit := util.FormatSize(pvc.AllocatedMB, pvc.UsedMB)

				report.WriteString(fmt.Sprintf("| %-15s | %-33s | %7.2f %-2s | %6.2f %-2s | %6.2f %-2s | %7d %% | %10d %% | %-15s |\n",
					nsReport.Namespace,
					pvc.Name,
					allocVal, unit,
					usedVal, unit,
					wastedVal, unit,
					pvc.UsedPct,
					pvc.WastagePct,
					ColorizeCategory(pvc.Category),
				))

				count++
				if count >= 5 {
					break
				}
			}
		}
		if count >= 5 {
			break
		}
	}

	report.WriteString(fmt.Sprintf("\n📄 Detailed CSV Report: %s\n", clusterReport.CSVFilePath))
	report.WriteString("─────────────────────────────────────────────\n")
	report.WriteString("✅ Audit completed successfully.\n")

	return report.String()
}

var pushgatewayServer string

var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Audit PVCs and generate wastage report",
	RunE: func(cmd *cobra.Command, args []string) error {

		// Determine namespaces
		var namespaces []string
		if allNamespaces {
			nsList, err := Internal.ListNamespaces()
			if err != nil {
				return err
			}
			namespaces = nsList
		} else {
			namespaces = []string{namespace}
		}

		clusterName := Internal.GetClusterName()
		clientset, config, err := Internal.GetK8sClientWithConfig()
		if err != nil {
			return err
		}

		var namespaceReports []NamespaceReport
		var csvRows [][]string
		csvRows = append(csvRows, []string{"Namespace", "PVC Name", "Allocated", "Used", "Wasted", "Used(%)", "Wastage(%)", "Attached Pod", "Category"})

		var highWastagePVCs, unattachedPVCs, cleanupCandidates []PVCInfo
		var totalPVCs, totalNamespaces int
		var totalAllocatedMB, totalUsedMB, totalWastedMB int64

		for _, ns := range namespaces {
			pvcs, err := Internal.ListPVCs(ns)
			if err != nil {
				fmt.Printf("Error listing PVCs in namespace %s: %v\n", ns, err)
				continue
			}
			if len(pvcs) == 0 {
				continue
			}

			nsReport := NamespaceReport{Namespace: ns}
			for _, pvc := range pvcs {
				allocated := pvc.Status.Capacity.Storage().Value() / 1024 / 1024 // MB

				podName, _, err := Internal.FindPodAndMountPathForPVC(ns, pvc.Name)
				var attachedPod string
				if err != nil || podName == "" {
					attachedPod = ""
					unattachedPVCs = append(unattachedPVCs, PVCInfo{Name: pvc.Name, Namespace: ns})
				} else {
					attachedPod = podName
				}

				// Get used size
				var usedMB int64
				if attachedPod != "" {
					usedMB, _ = Internal.GetUsedSizeInMB(clientset, config, ns, pvc.Name)
				}

				wastedMB := allocated - usedMB
				wastagePct := int64(0)
				usedPct := int64(0)
				if allocated > 0 {
					wastagePct = wastedMB * 100 / allocated
					usedPct = usedMB * 100 / allocated
				}

				// Assign category
				category := "Healthy"

				if attachedPod == "" && usedPct <= 5 {
					category = "Unused"

				} else if wastagePct == 100 {
					category = "Unused"
				} else if wastagePct <= 10 {
					category = "Critical"
				} else if wastagePct >= 70 {
					category = "Over-provisioned"
				}

				allocatedVal, allocatedUnit := util.FormatSizeMBorGB(allocated)
				usedVal, usedUnit := util.FormatSizeMBorGB(usedMB)
				wastedVal, wastedUnit := util.FormatSizeMBorGB(wastedMB)

				pvcInfo := PVCInfo{
					Name:          pvc.Name,
					Namespace:     ns,
					AllocatedMB:   allocated,
					Allocated:     allocatedVal,
					AllocatedUnit: allocatedUnit,
					UsedMB:        usedMB,
					Used:          usedVal,
					UsedUnit:      usedUnit,
					WastedMB:      wastedMB,
					Wasted:        wastedVal,
					WastedUnit:    wastedUnit,
					WastagePct:    int(wastagePct),
					UsedPct:       int64(usedPct),
					AttachedPod:   attachedPod,
					Category:      category,
				}

				nsReport.PVCs = append(nsReport.PVCs, pvcInfo)

				csvRows = append(csvRows, []string{
					ns,
					pvc.Name,
					fmt.Sprintf("%.2f %s", allocatedVal, allocatedUnit),
					fmt.Sprintf("%.2f %s", usedVal, usedUnit),
					fmt.Sprintf("%.2f %s", wastedVal, wastedUnit),
					fmt.Sprintf("%d", usedPct),
					fmt.Sprintf("%d", wastagePct),
					attachedPod,
					category,
				})

				if wastagePct > 80 {
					highWastagePVCs = append(highWastagePVCs, pvcInfo)
					cleanupCandidates = append(cleanupCandidates, pvcInfo)
				}

				totalAllocatedMB += allocated
				totalUsedMB += usedMB
				totalWastedMB += wastedMB
				totalPVCs++
			}

			if len(nsReport.PVCs) > 0 {
				namespaceReports = append(namespaceReports, nsReport)
				totalNamespaces++
			}
		}

		// Write CSV by default
		os.MkdirAll("reports", 0755)
		csvFile := filepath.Join("reports", fmt.Sprintf("pvc-wastage-report-%s.csv", time.Now().Format("20060102-150405")))
		file, err := os.Create(csvFile)
		if err != nil {
			return err
		}
		defer file.Close()
		writer := csv.NewWriter(file)
		defer writer.Flush()
		writer.WriteAll(csvRows)

		// Generate cluster report data
		totalAllocatedGB := float64(totalAllocatedMB) / 1024
		totalUsedGB := float64(totalUsedMB) / 1024
		totalWastedGB := float64(totalWastedMB) / 1024
		totalWastagePct := int64(0)
		if totalAllocatedMB > 0 {
			totalWastagePct = totalWastedMB * 100 / totalAllocatedMB
		}

		clusterReport := ClusterReport{
			ClusterName:        clusterName,
			GeneratedAt:        time.Now().Format("2006-01-02 15:04:05"),
			TotalNamespaces:    totalNamespaces,
			TotalPVCs:          totalPVCs,
			PVCsWithWastage:    len(highWastagePVCs),
			PVCsWithoutWastage: totalPVCs - len(highWastagePVCs),
			TotalAllocatedGB:   totalAllocatedGB,
			TotalUsedGB:        totalUsedGB,
			TotalWastedGB:      totalWastedGB,
			TotalWastagePct:    totalWastagePct,
			NamespaceReports:   namespaceReports,
			HighWastagePVCs:    highWastagePVCs,
			UnattachedPVCs:     unattachedPVCs,
			CleanupCandidates:  cleanupCandidates,
			CSVFilePath:        csvFile,
		}

		// Output
		if allNamespaces {
			if pushgatewayServer == "" {
				fmt.Println("One can provide --server-ip to push data to PushGateway when using --all-namespaces")
			}
			if pushgatewayServer != "" {
				err := PushPVCMetrics(pushgatewayServer, clusterReport)
				if err != nil {
					fmt.Printf("❌ Error pushing metrics: %v\n", err)
				} else {
					fmt.Printf("✅ Metrics pushed to Pushgateway at %s\n", pushgatewayServer)
				}
			}
			fmt.Println(GenerateCLIAuditReport(clusterReport))

		} else {
			PrintClusterReportCLI(clusterReport)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(auditCmd)
	auditCmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Kubernetes namespace")
	auditCmd.Flags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "Audit all namespaces")
	auditCmd.Flags().StringVarP(&pushgatewayServer, "server-ip", "s", "", "Pushgateway server IP (e.g., http://localhost:9091)")
}
