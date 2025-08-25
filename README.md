# 📂 Project Structure

### **`analyze.go`**

Implements the **`analyze`** command with subcommands:

* **`analyze cluster`** → Summarizes cluster-wide resources (total nodes, pods, storage classes).
* **`analyze nodes`** → Displays per-node allocatable CPU, memory, and storage.

---

### **`monitor.go`**

Implements the **`monitor`** command:

* Continuously queries **Prometheus** for pod/container CPU and memory usage.
* Compares **current usage vs. requests** against configurable thresholds.
* Generates **resize recommendations** (increase or decrease).
* Supports `--auto-apply` flag (currently simulated, does not modify workloads).

---

### **`pvc.go`**

Implements the **`pvc`** command with subcommand:

* **`pvc overprovisioned`** → Scans all PersistentVolumeClaims and detects wasted storage.
* Uses annotation **`resize-cltool/used-storage`** to compare **requested vs. actual usage**.

---

### **`root.go`**

Defines the **root CLI command** `resize-assistant`:

* Entry point for all subcommands (`analyze`, `monitor`, `pvc`, `simulate`).
* Provides global description and help menu.

---

### **`simulate.go`**

Implements the **`simulate`** command:

* Creates a **test namespace** (`resize-test`).
* Deploys workloads with different resource usage patterns:

  * CPU-intensive
  * Memory-hungry
  * Bursty
  * Underutilized
  * Balanced
* Creates PVCs with simulated usage annotations:

  * Overprovisioned
  * Well-sized
  * Underprovisioned
* Supports `--cleanup` to remove all simulation resources.

---

### **`utils.go`**

Provides **utility functions**:

* **`getClientSet`** → Creates a Kubernetes clientset:

  * Uses **in-cluster config** (when running inside Kubernetes).
  * Falls back to **local kubeconfig** when available.
* Shared helper across all commands for Kubernetes API access.

---