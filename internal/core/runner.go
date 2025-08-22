package core

import (
	"blockci-q/internal/storage"
	"fmt"
	"time"
)

// Runner ties together Parser + Scheduler + Executor
type Runner struct {
	Scheduler *Scheduler
	Executor  *Executor
	LogStorage *storage.LogStorage
}


func NewRunner() *Runner {
	return &Runner{
		Scheduler: NewScheduler(),
		Executor: NewExecutor(),
		LogStorage: storage.NewLogStorage("./logs"), // logs directory
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


			if err != nil {
				fmt.Printf("Step failed : %v\n", err)
				return err  // stop pipeline on failure
			}
			fmt.Println("  ✔ Step completed sucessfully")
		}
	}

	fmt.Println("\nPipeline finshed successfully ")
	return  nil
}


