package core

import (
	"blockci-q/internal/blockchain"
	"blockci-q/internal/storage"
	"blockci-q/pkg/utils"
	"fmt"
	"time"
)

// Runner ties together Parser + Scheduler + Executor + storage + blockchain
type Runner struct {
	Scheduler *Scheduler
	Executor  *Executor
	LogStorage *storage.LogStorage
	Ledger     *blockchain.Ledger
	AgentID		string // optional: identify which agent executed the step
}


func NewRunner() *Runner {

	// Initialize blockchain ledger (append only file)
	ledger, err := blockchain.OpenLedger("./ledger.json")
	if err != nil {
		// fail-open : you can choose to return error instead
		fmt.Printf("WARN : cannot open ledger: %v\n", err)
	}
	return &Runner{
		Scheduler: NewScheduler(),
		Executor: NewExecutor(),
		LogStorage: storage.NewLogStorage("./logs"), // logs directory
		Ledger: ledger,
		AgentID: "local-agent",  // replace if you have dynamic ids
	}
}


// RunPipeline executes all stages sequentally
func (r *Runner) RunPipeline(pipeline *Pipeline) error {
	fmt.Printf("Starting pipeline on agent : %s\n", pipeline.Agent)

	for i , stage := range pipeline.Stages {
		fmt.Printf("\n==> Stage %d :  %s\n" , i+1, stage.Name)

		steps := r.Scheduler.GetNextSteps(pipeline, i)
		for _, step := range steps {
			fmt.Printf("Running step : %s\n" , step.Run)

			output, err := r.Executor.RunStep(step, 5*time.Minute)
			fmt.Println("Output: \n", output)

			// Save log
			logPath ,logErr := r.LogStorage.SaveLog(stage.Name, step.Run, output)
			if logErr != nil {
				fmt.Printf("Failedto save logs: %v\n", logErr)
			} else {
				fmt.Printf("Log saved at : %s\n", logPath)
			}
			// Append blockchain blocks (best-effort ; do not block pipeline if ledger missing)
			if r.Ledger != nil && logErr == nil {


				logHash, hErr := utils.HashFile(logPath)
				if hErr != nil {
					fmt.Printf("WARN : cannot hash log : %v\n",hErr)
				}else {
					prev := r.Ledger.LastHash()
					idx := r.Ledger.NextIndex()
						blk , bErr := blockchain.NewBlock(idx,stage.Name,step.Run,logPath,logHash,prev,r.AgentID)
						if bErr != nil{
							fmt.Printf("WARN : cannot create block : %v\n",bErr)
						} else if aErr := r.Ledger.AppendBlocks(blk); aErr != nil {
							fmt.Printf("WARN : cannot append block : %v\n",aErr)
						}else{
							fmt.Printf("Ledger : appended blocks %d (hash=%s)\n",blk.Index,blk.Hash[:16])
						}			
				}
			}

			if err != nil {
				fmt.Printf("Step failed : %v\n", err)
				return err  // stop pipeline on failure
			}
			fmt.Println("  ✔ Step completed sucessfully")
		}
	}

	// Optional : verify the chain ath the end (cost some time on big ledgers)
	if r.Ledger != nil {
		if err := r.Ledger.VerifyChain(); err != nil {
			fmt.Printf("Ledger verification FAILED: %v\n", err)
		}else{
			fmt.Println("Ledger verification : ok")
		}
	}
	fmt.Println("\nPipeline finshed successfully ")
	return  nil
}


