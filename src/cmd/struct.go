package cmd

// PVCInfo stores detailed information about a single PVC
type PVCInfo struct {
	Name          string  // PVC name
	Namespace     string  // Namespace
	AllocatedMB   int64   // Allocated storage in MB
	Allocated     float64 // Allocated storage for display (MB or GB)
	AllocatedUnit string  // "MB" or "GB"
	UsedMB        int64   // Used storage in MB
	Used          float64 // Used storage for display
	UsedUnit      string  // "MB" or "GB"
	WastedMB      int64   // Wasted storage in MB
	Wasted        float64 // Wasted storage for display
	WastedUnit    string  // "MB" or "GB"
	WastagePct    int     // Percentage wasted
	AttachedPod   string  // Pod using the PVC (empty if unattached)
	Category      string
	Attached      bool
	UsedPct       int64
}

// NamespaceReport aggregates PVCs for a namespace
type NamespaceReport struct {
	Namespace string    // Namespace name
	PVCs      []PVCInfo // List of PVCs in this namespace
}

// ClusterReport aggregates all namespaces for a cluster
type ClusterReport struct {
	ClusterName        string            // Cluster name
	GeneratedAt        string            // Timestamp
	TotalNamespaces    int               // Count of namespaces audited
	TotalPVCs          int               // Count of PVCs audited
	PVCsWithWastage    int               // Number of PVCs with wastage > threshold
	PVCsWithoutWastage int               // PVCs without wastage
	TotalAllocatedGB   float64           // Total allocated storage in GB
	TotalUsedGB        float64           // Total used storage in GB
	TotalWastedGB      float64           // Total wasted storage in GB
	TotalWastagePct    int64             // Total cluster wastage percentage
	NamespaceReports   []NamespaceReport // Per-namespace details
	HighWastagePVCs    []PVCInfo         // PVCs with wastage > 80%
	UnattachedPVCs     []PVCInfo         // PVCs not attached to any pod
	CleanupCandidates  []PVCInfo         // Suggested PVCs for cleanup
	CSVFilePath        string            // Path to generated CSV file
}
