#make tidy      # Clean up go.mod/go.sum inside src/
#make build     # Build binary from src/
#make run       # Run CLI with default namespace
#make test      # Run all tests in src/
#make install   # Install binary globally
#make clean     # Cleanup
#make lint      # runs go vet + golangci-lint (if available)


.PHONY: build run docker clean tidy install test lint

BINARY_NAME = pvc-audit
SRC_DIR = src

# Build the binary from src/
build:
	@echo "🔨 Building $(BINARY_NAME)..."
	cd $(SRC_DIR) && go build -o ../$(BINARY_NAME) main.go

# Run the tool with a default command
run: build
	@echo "🚀 Running $(BINARY_NAME)..."
	./$(BINARY_NAME) list --namespace default

# Build Docker image
docker:
	@echo "🐳 Building Docker image..."
	docker build -t pvc-audit:latest .

# Cleanup build artifacts
clean:
	@echo "🧹 Cleaning up..."
	rm -f $(BINARY_NAME)

# Tidy Go modules (inside src/)
tidy:
	@echo "📦 Tidying Go modules..."
	cd $(SRC_DIR) && go mod tidy

# Run all Go tests (inside src/)
test:
	@echo "🧪 Running tests..."
	cd $(SRC_DIR) && go test ./...

# Lint the code (uses go vet, falls back to golangci-lint if available)
lint:
	@echo "🔎 Running go vet..."
	cd $(SRC_DIR) && go vet ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "✅ Running golangci-lint..."; \
		cd $(SRC_DIR) && golangci-lint run; \
	else \
		echo "⚠️  golangci-lint not installed, only ran go vet"; \
	fi

# Install the binary to $GOPATH/bin or $HOME/go/bin
install: build
	@echo "📥 Installing $(BINARY_NAME) to $$GOPATH/bin (or ~/go/bin)..."
	@mkdir -p $${GOPATH:-$$HOME/go}/bin
	cp $(BINARY_NAME) $${GOPATH:-$$HOME/go}/bin/
	@echo "✅ Installed. Make sure $${GOPATH:-$$HOME/go}/bin is in your PATH."
