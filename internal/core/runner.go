package core

import (
	"blockci-q/internal/security"
	"blockci-q/internal/storage"
	"blockci-q/pkg/utils"
	"crypto/ed25519"
	"fmt"
	"time"
)

// Runner ties together Parser + Scheduler + Executor + storage
// (agent no longer appends to ledger)
type Runner struct {
	Scheduler  *Scheduler
	Executor   *Executor
	LogStorage *storage.LogStorage
	PrivKey    ed25519.PrivateKey
	PubKey     ed25519.PublicKey
	AgentID    string
}

func NewRunner() *Runner {
	pub, priv, err := security.GenerateKeyPair()
	if err != nil {
		panic(err)
	}

	return &Runner{
		Scheduler:  NewScheduler(),
		Executor:   NewExecutor(),
		LogStorage: storage.NewLogStorage("./logs"),
		PrivKey:    priv,
		PubKey:     pub,
		AgentID:    "agent-1",
	}
}

// RunPipeline executes all stages sequentially
// Returns: map[stepID] → logPath
func (r *Runner) RunPipeline(pipeline *Pipeline) (map[string]string, error) {
	fmt.Printf("Starting pipeline on agent: %s\n", pipeline.Agent)

	results := make(map[string]string)

	for i, stage := range pipeline.Stages {
		fmt.Printf("\n==> Stage %d: %s\n", i+1, stage.Name)

		steps := r.Scheduler.GetNextSteps(pipeline, i)
		for _, step := range steps {
			stepID := fmt.Sprintf("%s/%s", stage.Name, step.Name)
			fmt.Printf("Running step: %s\n", stepID)

			// Run the actual step
			output, err := r.Executor.RunStep(step, 5*time.Minute)
			fmt.Println("Output:\n", output)

			// Save log locally
			logPath, logErr := r.LogStorage.SaveLog(stage.Name, step.Name, output)
			if logErr != nil {
				fmt.Printf("⚠️ Failed to save logs: %v\n", logErr)
			} else {
				fmt.Printf("Log saved at: %s\n", logPath)
				results[stepID] = logPath
			}

			if err != nil {
				fmt.Printf("❌ Step failed: %v\n", err)
				return results, err // stop pipeline on failure
			}
			fmt.Println("  ✔ Step completed successfully")
		}
	}

	fmt.Println("\nPipeline finished successfully (agent-side)")
	return results, nil
}

func ComputeLogHash(path string)(string, error){
	return utils.HashFile(path)
}
