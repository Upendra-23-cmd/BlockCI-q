package core

 // Schedular decides execution order of stages and jobs
 type Scheduler struct {}

 // NewScheduler creates a new scheduler
 func NewScheduler () *Scheduler{
	return &Scheduler{}
 }

 // GetNextJobs returns jobs for the current stage
 func (s *Scheduler) GetNextSteps(pipeline *Pipeline, stageIndex int) []Step {
	if stageIndex >= len(pipeline.Stages) {
		return nil
	}
	return pipeline.Stages[stageIndex].Steps
 }