BINARY_NAME=pvc-audit

.PHONY: build run docker

build:
	go build -o $(BINARY_NAME) main.go

run: build
	./$(BINARY_NAME) list --namespace default

docker:
	docker build -t pvc-audit:latest .
