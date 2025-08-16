package main

import (
	"blockci-q/internal/core"
	"log"
)


func main() {
	// Load pipeline.yaml
	pipeline, err := core.LoadPipeline("pipline.yaml")
	if err != nil {
		log.Fatalf("failed to load pipeline: %v", err)
	}

	// Run pipeline
	runner := core.NewRunner()
	if err := runner.RunPipeline(pipeline); err != nil {
		log.Fatalf("pipeline failed: %v", err)
	}
}