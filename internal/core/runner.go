package core

import (
	"blockci-q/internal/blockchain"
	"blockci-q/internal/security"
	"blockci-q/internal/storage"
	"crypto/ed25519"
	"fmt"
	"time"
)

// Runner ties together Parser + Scheduler + Executor + storage + blockchain
type Runner struct {
	Scheduler  *Scheduler
	Executor   *Executor
	LogStorage *storage.LogStorage
	Ledger     *blockchain.Ledger
	PrivKey    ed25519.PrivateKey
	PubKey     ed25519.PublicKey
	AgentID    string
}

func NewRunner() *Runner {
	pub, priv, err := security.GenerateKeyPair()
	if err != nil {
		panic(err)
	}

	// Initialize blockchain ledger (append only file)
	ledger, err := blockchain.OpenLedger("./ledger.json")
	if err != nil {
		fmt.Printf("WARN: cannot open ledger: %v\n", err)
	}

	return &Runner{
		Scheduler:  NewScheduler(),
		Executor:   NewExecutor(),
		LogStorage: storage.NewLogStorage("./logs"),
		Ledger:     ledger,
		PrivKey:    priv,
		PubKey:     pub,
		AgentID:    "agent-1",
	}
}

// RunPipeline executes all stages sequentially and returns a map of step -> logPath
func (r *Runner) RunPipeline(pipeline *Pipeline) (map[string]string, error) {
	fmt.Printf("Starting pipeline on agent: %s\n", pipeline.Agent)

	stepLogs := make(map[string]string)

	for i, stage := range pipeline.Stages {
		fmt.Printf("\n==> Stage %d : %s\n", i+1, stage.Name)

		steps := r.Scheduler.GetNextSteps(pipeline, i)
		for _, step := range steps {
			fmt.Printf("Running step: %s\n", step.Run)

			output, err := r.Executor.RunStep(step, 5*time.Minute)
			fmt.Println("Output:\n", output)

			// Save log
			logPath, logErr := r.LogStorage.SaveLog(stage.Name, step.Run, output)
			if logErr != nil {
				fmt.Printf("❌ Failed to save logs: %v\n", logErr)
			} else {
				fmt.Printf("Log saved at: %s\n", logPath)
				stepLogs[step.Run] = logPath
			}

			// Append to blockchain ledger
			// if r.Ledger != nil && logErr == nil {
			// 	logHash, hErr := utils.HashFile(logPath)
			// 	if hErr != nil {
			// 		fmt.Printf("⚠️ WARN: cannot hash log: %v\n", hErr)
			// 	} else {
			// 		prev := r.Ledger.LastHash()
			// 		idx := r.Ledger.NextIndex()
			// 		blk, bErr := blockchain.NewBlock(idx, stage.Name, step.Run, logPath, logHash, prev, r.AgentID)
			// 		if bErr != nil {
			// 			fmt.Printf("⚠️ WARN: cannot create block: %v\n", bErr)
			// 		}else {
			// 			fmt.Printf("✅ Ledger: appended block %d (hash=%s)\n", blk.Index, blk.Hash[:16])
			// 		}
			// 	}
			// }

			if err != nil {
				fmt.Printf("❌ Step failed: %v\n", err)
				return stepLogs, err
			}
			fmt.Println("  ✔ Step completed successfully")
		}
	}

	// Verify chain at the end
	if r.Ledger != nil {
		if err := r.Ledger.VerifyChain(); err != nil {
			fmt.Printf("❌ Ledger verification FAILED: %v\n", err)
		} else {
			fmt.Println("✅ Ledger verification: ok")
		}
	}

	fmt.Println("\nPipeline finished successfully")
	return stepLogs, nil
}
