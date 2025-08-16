package core

import (
	"fmt"
	"time"
)

// Runner ties together Parser + Scheduler + Executor
type Runner struct {
	Scheduler *Scheduler
	Executor  *Executor
}


func NewRunner() *Runner {
	return &Runner{
		Scheduler: NewScheduler(),
		Executor: NewExecutor(),
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


