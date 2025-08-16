package core

import (
	"bytes"
	"context"
	"os/exec"
	"time"
)

// Exector is responsible for runnig steps(commands)
type Executor struct {}

func NewExecutor() *Executor {
	return &Executor{}
}

// RunStepexecutes a single pipleine step and return its output+error
func (e *Executor) RunStep(step Step, timeout time.Duration) (string , error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Run the step in a shell (sh -c "cmd")
	cmd := exec.CommandContext(ctx, "sh" , "-c", step.Run)

	var out bytes.Buffer
	cmd.Stdout =  &out
	cmd.Stderr = &out

	err := cmd.Run()
	return out.String(), err
} 
