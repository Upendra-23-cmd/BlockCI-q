package core

// Job represents a single unit of work dispatched to an Agent.
// It is derived from a Pipeline's Stage+Step.
type Job struct {
	ID      string `json:"id"`      // unique job ID (server assigns, e.g., job-1)
	Stage   string `json:"stage"`   // stage name (e.g., "Build")
	Step    string `json:"step"`    // step name (e.g., "Compile")
	Cmd     string `json:"cmd"`     // command to run (from step.Run)
	AgentID string `json:"agentId"` // which agent this job is assigned to
	Status  string `json:"status"`  // pending, running, done, failed
}
