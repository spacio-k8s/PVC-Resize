# Contributing Guide for `pvc-audit`


🎉 First off, thanks for considering contributing!  

We aim to keep pvc-auditor welcoming and beginner-friendly — perfect for students, hobbyists, and professionals alike.

This guide explains how to set up your environment, commit changes, and follow project standards.


## Requirements
1. All commits must be signed using `git commit -s` (see DCO.txt).
2. By submitting contributions, you agree to the Contributor License Agreement (CLA.md).
3. Contributions will be licensed under the Apache 2.0 License (see LICENSE).

## Steps
- Fork the repository
- Create a feature branch
- Make your changes
- Commit with sign-off: `git commit -s -m "message"`
- Submit a Pull Request


## 🛠 Local Development Setup
### Prerequisites
- Go 1.21+
- Docker & kind/minikube (for local cluster testing)

### Getting Started

* Fork the repository and clone it locally:

  ```bash
  git clone https://github.com/spacio-k8s/PVCAuditor.git
  cd pvc-audit
  
  # Build CLI
  go mod tidy
   go build -o pvc-auditor main.go
   
  #Good to Run Auditor
  ./pvc-audit list 
  ```

###  Branching

* Always work on a feature branch (not `main`).
* Branch naming convention:

  * `feat/<feature-name>` → for new features
  * `fix/<bug-name>` → for bug fixes
  * `docs/<update>` → for documentation changes
  * `chore/<update>` → for cleanup, refactor, or dependency update

Example:
```bash
   git checkout -b feat/add-grafana-export
```


### Commit Messages

We use **conventional commits** to keep history clean:

* **feat:** → new feature
* **fix:** → bug fix
* **docs:** → documentation only
* **refactor:** → code change that doesn’t add features/fixes bugs
* **test:** → adding or updating tests
* **chore:** → tooling, build, dependency updates

**Format:**

```
<type>(scope): short description
```

**Examples:**

* `feat(cli): add --all-namespaces flag to list command`
* `fix(audit): correct calculation of wastage percentage`
* `docs: update usage examples in README`


## 4. Code Standards

* Follow Go best practices (`gofmt`, `golangci-lint`).
* Write modular, testable functions.
* Add **unit tests** for new logic where possible.
* Keep CLI help text and usage examples up to date.


### Testing Changes

Before pushing:

```bash
go fmt ./...
go vet ./...
go test ./... -v
```


### Submitting Changes

1. Commit with signed-off messages (required for DCO compliance):

   ```bash
   git commit -s -m "feat(audit): <ISSUE-ID> add prometheus pushgateway integration"
   ```

   > `-s` adds a `Signed-off-by` line to indicate you are the author.
2. Push your branch:

   ```bash
   git push origin feat/add-grafana-export
   ```
3. Open a Pull Request (PR) on GitHub.

### Pull Requests

- Fork the repo and create a new branch (feature/my-feature)

- Write clean, tested code

- Run go fmt ./... and golangci-lint run

- Open a PR with a clear description of your change


### PR Review Process

* At least one maintainer must approve before merging.
* All tests and CI checks must pass.
* Squash commits if needed for clean history.


### License and DCO

* By contributing, you agree your work will be licensed under the project’s open-source license.
* All commits must be signed (`git commit -s`).


###  Labels

- good first issue → great for newcomers

- help wanted → needs community help

- enhancement → feature requests

##  Code of Conduct

This project follows a Code of Conduct.

By participating, you agree to keep it respectful and inclusive.