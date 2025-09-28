# 🗂️ Spacio: PVC-Auditor  

`pvc-auditor` is a Kubernetes CLI tool that **tracks PersistentVolumeClaim (PVC) usage and wastage**.  It helps platform teams, SREs, and developers **identify unused or underutilized storage**, reducing costs and improving efficiency.  


## 🌍 Context  
Kubernetes makes it easy to allocate storage, but hard to track how much is actually used. Over time, this leads to:  
- PVCs that are **over-provisioned** (e.g. 500 Gi allocated, only 20 Gi used)  
- PVCs that are **unused or orphaned** (left behind after workload deletion)  
- PVCs that are **expensive** to maintain across cloud providers  

For platform teams, this creates **hidden costs, wasted resources, and operational risk**.  

**PVC-Auditor solves this by auditing your cluster’s PVCs and generating reports** with clear visibility into storage usage, waste, and savings potential.  

## 🚀 Why PVC-Auditor?  

- **Visibility first** → know what’s allocated vs. actually used  
- **Safe by design** → auditing only, no risky shrinking in OSS (shrinking in SaaS)  
- **Cloud agnostic** → works with AWS EBS, GCP PD, Azure Disk, and any CSI driver  
- **Contributor-friendly** → simple Go/Python codebase, great for students & OSS devs  
- **Path to automation** → start with audits, upgrade to SaaS for policies + shrinking  


## ⚖️ Open Source vs Enterprise  

💼 Enterprise Contact: For SaaS inquiries, pricing, or support, reach out at pvcauditor@gmail.com

| Feature | Open Source (CLI) | Enterprise SaaS |
|---------|------------------|-----------------|
| **Quick Setup** | ✅ Lightweight CLI & agent | ✅ Same simplicity, SaaS dashboard |
| **PVC Discovery** | ✅ Scan all PVCs in a cluster | ✅ Multi-cluster discovery |
| **Usage vs Allocation Reports** | ✅ Text-based reports | ✅ Rich dashboards & analytics |
| **Wastage Detection** | ✅ Find unused, orphaned, over-provisioned PVCs | ✅ Automated cleanup + alerts |
| **Cluster Scope** | ✅ Single-cluster only | 🌐 Multi-cluster visibility |
| **Automation** | ❌ Manual review only | 🤖  Auto-managed PVCs (resize + cleanup)|
| **Governance** | ❌ Not included | 🔒 RBAC, approvals, policy enforcement |
| **Integrations** | ❌ Not included | 🔔 Slack, Jira, Cloud cost tools |


👉 Think of **PVC-Auditor OSS** as your **storage magnifying glass**

👉 And **Enterprise SaaS** as your **cost-optimization autopilot** — PVCs are automatically managed, resized, and cleaned up without manual effort.

## ✨ Features (Open Source)  

- ⚡ **Quick Setup** — lightweight CLI & agent  
- 🔍 **PVC Discovery** — scan all PVCs in a cluster  
- 📊 **Usage vs Allocation Reports** — see storage requests vs actual usage  
- 🗑️ **Wastage Detection** — find unused, orphaned, and over-provisioned volumes  
- 🎯 **Single-Cluster Focus** — MVP designed per cluster (multi-cluster in SaaS)  
- 🔗 **Pod Mapping** — see which PVCs are attached (or unattached) to pods
- 🛠 **Test Mode** — inject dummy usage data for demo & validation
- 📡**Prometheus Metrics** — export stats for Grafana dashboards


## 🔒 Enterprise SaaS Features  

Upgrade for **automation, scale, and governance**:  

- 🌐 **Multi-Cluster Management** — centralized visibility across clusters  
- 🤖 **Automated Shrinking** — safely resize PVCs to right-size capacity  
- 🛡️ **Policy-Driven Governance** — org-wide storage usage rules  
- ✅ **Approval Workflows** — integrate with DevOps/SRE review  
- 👥 **RBAC** — fine-grained team-based permissions  
- 📈 **Dashboards & Analytics** — visual insights for finance, ops & platform teams  
- 🔔 **Integrations** — Slack, Jira, and cloud cost dashboards  


## ⚡ Installation  

```bash
# Clone repo
git clone https://github.com/spacio-k8s/PVCAuditor.git
cd pvc-auditor

# Build CLI
    cd src
    go mod tidy
    go build -o pvc-auditor main.go
or 
    make tidy build

# Run audit
./pvc-auditor audit --all-namespace
```


## 📊 Example Report  

```bash
PVC Audit Summary Report
──────────────────────────────────────────────
Cluster Name           : K8sCluster-1.33
Generated At           : 2025-09-26 23:19:16
Total Namespaces Audited : 13
Total PVCs Audited       : 14

PVC Space Summary
──────────────────────────────────────────────
Total Allocated Space (GB) : 30.00 GB
Total Used Space (GB)      : 24.00 GB
Total Wasted Space (GB)    : 16 GB
Wastage Percentage         : 53.33%

PVC Wastage Details
──────────────────────────────────────────────
PVCs with High Wastage (≥80%) : 2
Unattached PVCs              : 10
Cleanup Candidates           : 2

Top 5 High Wastage PVCs
──────────────────────────────────────────────
< PVC Details in table as shown below> 
```
## Sample PVC Audit Report  :


| Namespace       | PVC Name          | Allocated  | Used     | Wasted   | Used (%) | Wastage (%) | Status        |
|-----------------|-----------------|-----------|----------|----------|----------|-------------|---------------|
| default         | prometheus       | 10 GB     | 0 GB     | 10 GB    | 0 %      | 100 %       | ![Unused](https://img.shields.io/badge/Unused-9e9e9e?style=flat-square&logo=kubernetes&logoColor=white) |
| pvc-test        | test-pvc-small   | 5 GB      | 1 GB     | 4 GB     | 20 %     | 80 %        | ![Over-provisioned](https://img.shields.io/badge/Over--provisioned-ff9800?style=flat-square&logo=kubernetes&logoColor=white) |
| pvc-critical    | app-cache        | 10 GB     | 9 GB     | 1 GB     | 90 %     | 10 %        | ![Critical](https://img.shields.io/badge/Critical-d32f2f?style=flat-square&logo=kubernetes&logoColor=white) |
| pvc-healthy     | db-storage       | 10 GB     | 6 GB    | 7 GB     |  60%     | 40 %         | ![Healthy](https://img.shields.io/badge/Healthy-4caf50?style=flat-square&logo=kubernetes&logoColor=white) |




### 📖 Legend  
-  **Over-provisioned** → PVC has far more allocated than used , more than 70%
- **Unused** → PVC allocated but not used at all (candidate for deletion)  
- **Critical** → PVC used nearly full (risk of outage)  



## 📊 Grafana Dashboard Integration

PVC-Auditor can export metrics for visualization in Grafana. Follow these steps to set it up:



### 1️⃣ Run PVC-Auditor with metrics export
```bash
./pvc-auditor audit --all-namespace --server-ip <IP>
```

* Replace `<IP>` with your Kubernetes API server or metrics endpoint.
* This will expose Prometheus-compatible metrics.

### 2️⃣ Import Dashboard into Grafana

1. Open Grafana → **Dashboards → Manage → Import**
2. Upload the downloaded JSON file for the PVC-Auditor dashboard ((Ref: PVC_Stats_Grafana.json)
3. Select the Prometheus data source used by PVC-Auditor
4. Click **Import**

### 3️⃣ View Metrics

The dashboard provides:

* PVC usage vs allocated space
* Wasted storage per namespace
* Unattached or orphaned PVCs
* Top over-provisioned PVCs


### Sample Grafana Dashboard 
![Grafana Dashboard](https://raw.githubusercontent.com/spacio-k8s/PVCAuditor/refs/heads/main/images/sample_grafana_dashboard.jpeg?token=GHSAT0AAAAAADL7XPZUJR6EI3J5DMVI5VY42GW52SQ)


# 🤝 Contributing

We welcome contributors of all experience levels 🙌

* Expand docs & tutorials
* Add features & bug fixes
* Improve CLI experience

📘 See CONTRIBUTING.md to get started.

## 📜 License

Licensed under [Apache 2.0](./LICENSE).

## 💼 Enterprise Contact

For Enterprise SaaS inquiries, pricing, or support, please contact us at: 📧 pvcauditor@gmail.com


👉 This README is user-facing, while `PVC_AUDIT_CLI_GUIDE.md` can be a **detailed reference** for all flags and examples.
