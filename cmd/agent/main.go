package main

import (
	"blockci-q/internal/core"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Agent struct {
	ID   string `json:"id"`
	Host string `json:"host"`
}

type Job struct {
	ID    string `json:"id"`
	Stage string `json:"stage"`
	Step  string `json:"step"`
	Cmd   string `json:"cmd"`
	Status string `json:"status"`
	AgentID string `json:"agentId"`
}

func main() {
	serverURL := "http://localhost:8080"
	agentID := "agent-1"

	// create one runner for the agent lifetime
	runner := core.NewRunner()

	if err := registerAgent(serverURL, agentID); err != nil {
		fmt.Println("❌ failed to register agent:", err)
		os.Exit(1)
	}

	fmt.Println("✅ Agent started, polling for jobs...")
	pollJobs(serverURL, agentID, runner)
}

// registerAgent registers this agent with the server
func registerAgent(serverURL, id string) error {
	agent := Agent{ID: id, Host: "localhost"}
	data, _ := json.Marshal(agent)

	resp, err := http.Post(serverURL+"/agent/register", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to register agent: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("register failed: %s", string(body))
	}

	fmt.Println("✅ Agent registered:", id)
	return nil
}

// pollJobs continuously polls the server for the next job
func pollJobs(serverURL, id string, runner *core.Runner) {
	for {
		url := fmt.Sprintf("%s/agents/%s/jobs/next", serverURL, id)
		resp, err := http.Get(url)
		if err != nil {
			fmt.Println("⚠️ poll error:", err)
			time.Sleep(5 * time.Second)
			continue
		}

		// Handle NoContent (no job)
		if resp.StatusCode == http.StatusNoContent {
			resp.Body.Close()
			time.Sleep(2 * time.Second)
			continue
		}

		// Expect 200 OK with job JSON
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			fmt.Printf("⚠️ unexpected status %d: %s\n", resp.StatusCode, string(body))
			time.Sleep(3 * time.Second)
			continue
		}

		var job Job
		if err := json.NewDecoder(resp.Body).Decode(&job); err != nil {
			fmt.Println("⚠️ decode job error:", err)
			resp.Body.Close()
			time.Sleep(2 * time.Second)
			continue
		}
		resp.Body.Close()

		fmt.Printf("📥 Received job: %s (cmd=%s)\n", job.ID, job.Cmd)

		// execute job using the runner
		output, success, logPath := runJob(job, id, runner)

		// report result back to server (server will append to ledger)
		reportResult(serverURL, job, id, output, success, logPath)
	}
}

// runJob executes the job using the provided runner.
// Returns: output human message, success flag, logPath (may be empty if log save failed)
func runJob(job Job, agentID string, runner *core.Runner) (string, bool, string) {
	// Build a minimal pipeline for this single-step job
	pipeline := &core.Pipeline{
		Agent: agentID,
		Stages: []core.Stage{
			{
				Name: job.Stage,
				Steps: []core.Step{
					{Run: job.Cmd},
				},
			},
		},
	}

	// Run pipeline; runner will save step logs and return map[stepCmd] -> logPath
	results, err := runner.RunPipeline(pipeline)
	if err != nil {
		return fmt.Sprintf("job %s failed: %v", job.ID, err), false, ""
	}

	// Extract the first (and expected only) logPath from results map
	var logPath string
	for _, lp := range results {
		logPath = lp
		break
	}

	return fmt.Sprintf("job %s completed successfully", job.ID), true, logPath
}

// reportResult sends a structured result JSON to the server so server can record it in ledger
func reportResult(serverURL string, job Job, agentID string, output string, success bool, logPath string) {
	result := map[string]interface{}{
		"id":      job.ID,
		"stage":   job.Stage,
		"step":    job.Step,
		"cmd":     job.Cmd,
		"agentID": agentID,
		"output":  output,
		"logPath": logPath,
		"success": success,
		"time":    time.Now().Format(time.RFC3339),
	}

	data, _ := json.Marshal(result)
	resp, err := http.Post(serverURL+"/jobs/result", "application/json", bytes.NewBuffer(data))
	if err != nil {
		fmt.Println("⚠️ failed to report result:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println("📤 Reported result:", string(body))
}
