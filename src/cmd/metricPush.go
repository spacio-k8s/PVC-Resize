package cmd

import (
	"fmt"
	"pvc-audit/util"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

func PrintClusterReportCLI(report ClusterReport) {
	fmt.Println("\n -----------------------------------------")
	fmt.Println("🎯📊 PVC Audit Summary Report - Cluster View 🎯")
	fmt.Println("-----------------------------------------")
	fmt.Printf("Cluster: %s\n", report.ClusterName)
	fmt.Printf("Generated At: %s\n\n", report.GeneratedAt)

	// Cluster summary
	fmt.Println("🧱 PVC Space Summary")
	fmt.Println("-----------------------------------------")
	fmt.Printf("Total Namespaces Audited: %d\n", report.TotalNamespaces)
	fmt.Printf("Total PVCs Audited: %d\n", report.TotalPVCs)
	fmt.Printf("PVCs with Wastage: %d\n", report.PVCsWithWastage)
	fmt.Printf("PVCs without Wastage: %d\n", report.PVCsWithoutWastage)
	fmt.Printf("Total Allocated (GB): %.2f\n", report.TotalAllocatedGB)
	fmt.Printf("Total Used (GB): %.2f\n", report.TotalUsedGB)
	fmt.Printf("Total Wasted (GB): %.2f (%.0f%%)\n\n",
		report.TotalWastedGB,
		float64(report.TotalWastedGB*100)/report.TotalAllocatedGB,
	)

	// Namespace-wise details
	for _, nsReport := range report.NamespaceReports {
		fmt.Printf("\n🔹 Namespace: %s\n", nsReport.Namespace)
		fmt.Println("--------------------------------------------------------------------------------------------------------------------------------------")
		fmt.Printf("%-25s %-10s %-15s %-15s %-10s %-15s %-12s %-20s\n",
			"PVC NAME", "ATTACHED", "ALLOCATED", "USED", "USED(%)", "WASTED", "WASTAGE(%)", "CATEGORY")
		fmt.Println("--------------------------------------------------------------------------------------------------------------------------------------")

		for _, pvc := range nsReport.PVCs {
			// attached check
			attached := "Yes"
			for _, u := range report.UnattachedPVCs {
				if u.Name == pvc.Name && u.Namespace == pvc.Namespace {
					attached = "No"
					break
				}
			}

			// calculate Used%
			usedPct := 0.0
			if pvc.AllocatedMB > 0 {
				usedPct = (float64(pvc.UsedMB) / float64(pvc.AllocatedMB)) * 100
			}

			// determine category
			category := "Healthy"
			if attached == "No" {
				category = "Unattached"
			} else if pvc.WastagePct >= 80 {
				category = "High Wastage"
			} else if pvc.WastedMB > 1024 { // >1 GB wasted
				category = "Cleanup Candidate"
			}

			// format sizes using your util
			allocVal, allocUnit := util.FormatSizeMBorGB(pvc.AllocatedMB)
			usedVal, usedUnit := util.FormatSizeMBorGB(pvc.UsedMB)
			wastedVal, wastedUnit := util.FormatSizeMBorGB(pvc.WastedMB)

			allocStr := fmt.Sprintf("%.2f %s", allocVal, allocUnit)
			usedStr := fmt.Sprintf("%.2f %s", usedVal, usedUnit)
			wastedStr := fmt.Sprintf("%.2f %s", wastedVal, wastedUnit)

			// print row
			fmt.Printf("%-25s %-10s %-15s %-15s %-10.1f %-15s %-12d %-20s\n",
				pvc.Name,
				attached,
				allocStr,
				usedStr,
				usedPct,
				wastedStr,
				pvc.WastagePct,
				category,
			)
		}
	}
}

func PushPVCMetrics(pushGateway string, clusterReport ClusterReport) error {
	cluster := clusterReport.ClusterName

	// Cluster-level metrics
	pushCollector := []prometheus.Collector{

		prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        "pvc_total_allocated_gb",
			Help:        "Total allocated PVC space in GB",
			ConstLabels: prometheus.Labels{"cluster": cluster},
		}),
		prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        "pvc_total_used_gb",
			Help:        "Total used PVC space in GB",
			ConstLabels: prometheus.Labels{"cluster": cluster},
		}),
		prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        "pvc_total_wasted_gb",
			Help:        "Total wasted PVC space in GB",
			ConstLabels: prometheus.Labels{"cluster": cluster},
		}),
		prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        "pvc_total_pvcs",
			Help:        "Total number of PVCs",
			ConstLabels: prometheus.Labels{"cluster": cluster},
		}),
		prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        "pvc_pvcs_with_wastage",
			Help:        "PVCs with high wastage",
			ConstLabels: prometheus.Labels{"cluster": cluster},
		}),
		prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        "pvc_unattached",
			Help:        "Number of unattached PVCs",
			ConstLabels: prometheus.Labels{"cluster": cluster},
		}),
		prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        "pvc_total_namespaces",
			Help:        "Total namespaces audited in the cluster",
			ConstLabels: prometheus.Labels{"cluster": cluster},
		}),
		prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        "pvc_cleanup_candidates",
			Help:        "Number of PVCs eligible for cleanup",
			ConstLabels: prometheus.Labels{"cluster": cluster},
		}),
	}

	// Set cluster-level values
	pushCollector[0].(prometheus.Gauge).Set(clusterReport.TotalAllocatedGB)
	pushCollector[1].(prometheus.Gauge).Set(clusterReport.TotalUsedGB)
	pushCollector[2].(prometheus.Gauge).Set(clusterReport.TotalWastedGB)
	pushCollector[3].(prometheus.Gauge).Set(float64(clusterReport.TotalPVCs))
	pushCollector[4].(prometheus.Gauge).Set(float64(clusterReport.PVCsWithWastage))
	pushCollector[5].(prometheus.Gauge).Set(float64(len(clusterReport.UnattachedPVCs)))
	pushCollector[6].(prometheus.Gauge).Set(float64(len(clusterReport.NamespaceReports)))
	pushCollector[7].(prometheus.Gauge).Set(float64(len(clusterReport.CleanupCandidates)))

	// Namespace-level metrics
	for _, nsReport := range clusterReport.NamespaceReports {
		ns := nsReport.Namespace
		var nsAllocatedGB, nsUsedGB, nsWastedGB float64
		var nsPVCsWithWastage int

		for _, pvc := range nsReport.PVCs {
			nsAllocatedGB += pvc.Allocated
			nsUsedGB += pvc.Used
			nsWastedGB += pvc.Wasted
			if pvc.WastagePct >= 80 {
				nsPVCsWithWastage++
			}

			// Per-PVC metrics
			pushCollector = append(pushCollector, prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "pvc_allocated_gb",
				Help: "PVC allocated GB",
				ConstLabels: prometheus.Labels{
					"cluster":   cluster,
					"namespace": ns,
					"pvc":       pvc.Name,
					"pod":       pvc.AttachedPod,
				},
			}))
			pushCollector[len(pushCollector)-1].(prometheus.Gauge).Set(pvc.Allocated)

			pushCollector = append(pushCollector, prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "pvc_used_gb",
				Help: "PVC used GB",
				ConstLabels: prometheus.Labels{
					"cluster":   cluster,
					"namespace": ns,
					"pvc":       pvc.Name,
					"pod":       pvc.AttachedPod,
				},
			}))
			pushCollector[len(pushCollector)-1].(prometheus.Gauge).Set(pvc.Used)

			pushCollector = append(pushCollector, prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "pvc_wasted_gb",
				Help: "PVC wasted GB",
				ConstLabels: prometheus.Labels{
					"cluster":   cluster,
					"namespace": ns,
					"pvc":       pvc.Name,
					"pod":       pvc.AttachedPod,
				},
			}))
			pushCollector[len(pushCollector)-1].(prometheus.Gauge).Set(pvc.Wasted)

			pushCollector = append(pushCollector, prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "pvc_wastage_pct",
				Help: "PVC wastage %",
				ConstLabels: prometheus.Labels{
					"cluster":   cluster,
					"namespace": ns,
					"pvc":       pvc.Name,
					"pod":       pvc.AttachedPod,
				},
			}))
			pushCollector[len(pushCollector)-1].(prometheus.Gauge).Set(float64(pvc.WastagePct))
		}

		// Namespace aggregated metrics
		pushCollector = append(pushCollector, prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        "pvc_namespace_allocated_gb",
			Help:        "Namespace allocated GB",
			ConstLabels: prometheus.Labels{"cluster": cluster, "namespace": ns},
		}))
		pushCollector[len(pushCollector)-1].(prometheus.Gauge).Set(nsAllocatedGB)

		pushCollector = append(pushCollector, prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        "pvc_namespace_used_gb",
			Help:        "Namespace used GB",
			ConstLabels: prometheus.Labels{"cluster": cluster, "namespace": ns},
		}))
		pushCollector[len(pushCollector)-1].(prometheus.Gauge).Set(nsUsedGB)

		pushCollector = append(pushCollector, prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        "pvc_namespace_wasted_gb",
			Help:        "Namespace wasted GB",
			ConstLabels: prometheus.Labels{"cluster": cluster, "namespace": ns},
		}))
		pushCollector[len(pushCollector)-1].(prometheus.Gauge).Set(nsWastedGB)

		pushCollector = append(pushCollector, prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        "pvc_namespace_pvcs_with_wastage",
			Help:        "Namespace PVCs with high wastage",
			ConstLabels: prometheus.Labels{"cluster": cluster, "namespace": ns},
		}))
		pushCollector[len(pushCollector)-1].(prometheus.Gauge).Set(float64(nsPVCsWithWastage))
	}

	// Push all metrics to Pushgateway
	pusher := push.New(pushGateway, "pvc_audit_metrics")
	for _, c := range pushCollector {
		pusher.Collector(c)
	}
	if err := pusher.Push(); err != nil {
		return fmt.Errorf("could not push metrics: %v", err)
	}

	fmt.Println("✅ Metrics pushed successfully to Pushgateway:", pushGateway)
	return nil
}
