# 🧩 PVC Audit CLI - Quick Start Guide

The **PVC Audit CLI** helps track Persistent Volume Claim (PVC) usage, wastage, and pods attached to PVCs in Kubernetes clusters. It also supports test data dumps and Prometheus metrics.

---

## 1️⃣ Audit Commands – Check PVC Usage & Wastage

| Command | Description |
|---------|-------------|
| `./pvc-audit audit` | 🎯 Audit PVCs in the **default namespace** and show wastage. |
| `./pvc-audit audit -n <namespace>` | 📌 Audit PVCs in a **specific namespace**. |
| `./pvc-audit audit --all-namespaces` | 🌐 Audit **all namespaces** (future). |
| `./pvc-audit audit -s <server-ip>` | 📊 Push metrics to **Prometheus Pushgateway**. |

**Flags:**
- `-A, --all-namespaces` – Audit all namespaces  
- `-n, --namespace string` – Specify namespace (default: `default`)  
- `-s, --server-ip string` – Push metrics to Prometheus Pushgateway  
- `-h, --help` – Show command help  

**Example:**
```bash
# Audit PVCs in namespace "pvc-test"
./pvc-audit audit -n pvc-test
````



## 2️⃣ List / Discovery Commands – Explore PVCs & Pods

| Command                             | Description                                                  |
| ----------------------------------- | ------------------------------------------------------------ |
| `./pvc-audit list`                  | 📋 List all PVCs in **default namespace**.                   |
| `./pvc-audit list -n <namespace>`   | 📌 List PVCs in a **specific namespace**.                    |
| `./pvc-audit list --all-namespaces` | 🌐 List PVCs across the **entire cluster**.                  |
| `./pvc-audit pods -n <namespace>`   | 🐳 List pods **attached/unattached** to PVCs in a namespace. |
| `./pvc-audit pods --all-namespaces` | 🌍 List pods **attached/unattached** in all namespaces.      |

**Example:**

```bash
# List PVCs and attached pods in "dev" namespace
./pvc-audit list -n dev
./pvc-audit pods -n dev
```


## 3️⃣ Dump / Test Commands – Simulate PVC Usage

| Command                                                     | Description                           |
| ----------------------------------------------------------- | ------------------------------------- |
| `./pvc-audit dump -n <namespace> -p <pvc-name> --size 100M` | 💾 Fill PVC with **dummy/test data**. |


**Example using BusyBox in a pod:**

```bash
kubectl exec -it pvc-waste-pod -n pvc-waste-test -- sh
cd /mnt/data
# Write 100MB dummy file
dd if=/dev/zero of=testfile bs=1M count=100
ls -lh testfile
du -sh testfile
```

**Flags:**

* `-A, --all-namespaces` – Operate on all namespaces
* `-n, --namespace string` – Specify namespace
* `-p, --pvc string` – PVC name
* `-s, --size string` – Fill PVC with **test data in MB**
* `-h, --help` – Show command help



## 4️⃣ Expose Metrics for Prometheus

Push PVC usage and wastage metrics for monitoring:

```bash
./pvc-audit audit -n pvc-test -s http://localhost:9091
```

**Tip:** Use this in combination with `--all-namespaces` to get cluster-wide metrics.



## 5️⃣ Recommended Workflow

1. **Discover PVCs**:

```bash
./pvc-audit list --all-namespaces
```

2. **Audit PVC usage/wastage**:

```bash
./pvc-audit audit -n dev
```

3. **Simulate data for testing**:

```bash
./pvc-audit dump -n dev -p pvc-demo --size 100M
```


4. **Expose metrics to Prometheus**:

```bash
./pvc-audit audit --all-namespaces -s http://pushgateway:9091
```


## 6️⃣ General Help

```bash
./pvc-audit --help        # Main CLI help
./pvc-audit <command> --help  # Help for a specific command
```



## ✅ Summary

* Audit PVC wastage and optionally push metrics to Prometheus.
* List PVCs and pods across namespace(s).
* Dump/cleanup test data to simulate PVC usage.
* Supports **namespace-scoped** or **cluster-wide** operations.


## ✅ Quick Tips

- Combine list, audit, and pods for full PVC insight.
- Use dump to simulate usage safely in dev/test.
- Use -s flag to integrate with Prometheus Pushgateway for metrics.
- --all-namespaces is useful for cluster-wide audits.