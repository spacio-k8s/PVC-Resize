# Contributing to pvc-auditor

🎉 First off, thanks for considering contributing!  

We aim to keep pvc-auditor welcoming and beginner-friendly — perfect for students, hobbyists, and professionals alike.

---

## 🛠 Local Development Setup

### Prerequisites
- Go 1.21+
- Docker & kind/minikube (for local cluster testing)

### Getting Started

```
# Fork and clone your fork
git clone <>
cd pvc-auditor

# Build CLI
go build -o pvc-auditor ./cmd

# Run an audit
./pvc-auditor audit --output json


# Go unit tests
go test ./...

# Python backend tests
pytest backend/tests

```


## Pull Requests

- Fork the repo and create a new branch (feature/my-feature)

- Write clean, tested code

- Run go fmt ./... and golangci-lint run

- Open a PR with a clear description of your change

##  Labels

- good first issue → great for newcomers

- help wanted → needs community help

- enhancement → feature requests

##  Code of Conduct

This project follows a Code of Conduct.

By participating, you agree to keep it respectful and inclusive.