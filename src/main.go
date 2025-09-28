package main

import (
	"log"
	"pvc-audit/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatalf("Error executing pvc-audit: %v", err)
	}
}
