#  Spacio : PVC-Auditor

**pvc-auditor** is an **open-source Kubernetes CLI and agent** that helps teams **find and understand PersistentVolumeClaim (PVC) waste** in their clusters.

Kubernetes makes it easy to allocate storage, but hard to track how much is actually used. Over time, this leads to:

- PVCs that are over-provisioned (e.g. 500 Gi allocated, only 20 Gi used)

- PVCs that are unused or orphaned (left behind after workload deletion)

- PVCs that are expensive to maintain across cloud providers

For platform teams, **this creates hidden costs, wasted resources, and operational risk.**

**pvc-auditor solves this problem by auditing your cluster’s PVCs and generating detailed reports:**

- CLI reports for automation pipelines

- JSON/YAML outputs for GitOps and integrations

- Rich HTML dashboards for engineers and managers to review

The result: **clear visibility into storage usage, waste, and savings potential.**

### Why PVC-Auditor?

- **Visibility first** → know what storage is allocated vs. actually used

- **Safe by design**  → auditing only, no risky shrinking logic in the OSS CLI

- **Cloud agnostic** → works on AWS EBS, GCP PD, Azure Disk, and any CSI driver

- **Contributor-friendly** → simple Go/Python codebase, great for students & OSS devs

- **Path to automation** → upgrade to the SaaS edition for shrinking, policies, and multi-cluster support

### How It Fits

✅ **Open Source** (this repo): Single-cluster audits, reports, developer contributions

🔒 **Enterprise SaaS**: Multi-cluster management, automated shrinking, approvals, RBAC, dashboards

Think of **pvc-auditor** as your first step toward **cost-aware Kubernetes storage management.**
Audit today. Shrink tomorrow. 🚀


## ✨ Features (Open Source)

- **Quick Setup** — lightweight CLI & agent  
-  **PVC Discovery** — scan all PVCs in a cluster  
- **Usage vs Allocation Reports** — output as Table, JSON, YAML, or HTML  
- **Wastage Detection** — unused, orphaned, and over-provisioned volumes  
-  **Cloud-Agnostic** — AWS EBS, GCP PD, Azure Disk, on-prem CSI  
- **Single-Cluster Focus** — MVP works per cluster (multi-cluster in SaaS)  



## Installation

```bash
# Clone repo
git clone <>
cd pvc-auditor

# Build CLI
go build -o pvc-auditor ./cmd

# Run audit
./pvc-auditor audit --output html

## Example Report

## Contributing

We welcome contributors of all experience levels 🙌

Ways you can help:

- Add support for new StorageClasses
- Improve HTML dashboards
- Add CLI flags for filtering/sorting
- Write tests for PVC scanning logic

Expand docs & tutorials

📘 See CONTRIBUTING.md to get started.

We use GitHub Issues for bugs/features and Discussions for roadmap ideas.
Look for good first issue and help wanted labels to dive in!