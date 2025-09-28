package main

import (
	"blockci-q/internal/core"
	"blockci-q/pkg/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
	"github.com/google/uuid"

)

type Agent struct {
	ID   string `json:"id"`
	Host string `json:"host"`
}

type Job struct {
	ID         string `json:"id"`
	Stage      string `json:"stage"`
	Step       string `json:"step"`
	Cmd        string `json:"cmd"`
	PipelineID string `json:"pipelineId"`
}

func main() {
	serverURL := "http://localhost:8080"

	// ✅ Dynamic Agent ID (via ENV or auto-generated)
	agentID := os.Getenv("AGENT_ID")
	if agentID == "" {
		agentID = fmt.Sprintf("agent-%s", uuid.New().String()[:8]) // short UUID
	}

	runner := core.NewRunner()

	if err := registerAgent(serverURL, agentID); err != nil {
		fmt.Println("❌ failed to register agent:", err)
		os.Exit(1)
	}

	fmt.Println("✅ Agent started with ID:", agentID)
	pollJobs(serverURL, agentID, runner)
}

// registerAgent registers the agent with the server
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

	fmt.Println("🤝 Agent registered with server:", id)
	return nil
}


// pollJobs keeps polling the server for new jobs
func pollJobs(serverURL, id string, runner *core.Runner) {
	for {
		url := fmt.Sprintf("%s/agents/%s/jobs/next", serverURL, id)
		resp, err := http.Get(url)
		if err != nil {
			fmt.Println("⚠️ poll error:", err)
			time.Sleep(5 * time.Second)
			continue
		}

		if resp.StatusCode == http.StatusNoContent {
			resp.Body.Close()
			time.Sleep(2 * time.Second)
			continue
		}

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

		fmt.Printf("📥 Received job: %s (cmd=%s, pipeline=%s)\n", job.ID, job.Cmd, job.PipelineID)

		// run job
		output, success, logPath, logHash := runJob(job, id, runner)

		// report result
		reportResult(serverURL, job, id, output, success, logPath, logHash)
	}
}

// runJob executes the job with the runner
func runJob(job Job, agentID string, runner *core.Runner) (string, bool, string, string) {
	// build pipeline for this job
	pipeline := &core.Pipeline{
		Agent: agentID,
		Stages: []core.Stage{
			{
				Name: job.Stage,
				Steps: []core.Step{
					{Name: job.Step, Run: job.Cmd},
				},
			},
		},
	}

	results, err := runner.RunPipeline(pipeline)
	if err != nil {
		return fmt.Sprintf("job %s failed: %v", job.ID, err), false, "", ""
	}

	// pick logPath
	var logPath string
	for _, lp := range results {
		logPath = lp
		break
	}

	// compute log hash
	logHash := ""
	if logPath != "" {
		h, err := utils.HashFile(logPath)
		if err == nil {
			logHash = h
		}
	}

	return fmt.Sprintf("job %s completed successfully", job.ID), true, logPath, logHash
}

// reportResult sends job results back to server
func reportResult(serverURL string, job Job, agentID string, output string, success bool, logPath, logHash string) {
	result := map[string]interface{}{
		"id":         job.ID,
		"stage":      job.Stage,
		"step":       job.Step,
		"cmd":        job.Cmd,
		"pipelineId": job.PipelineID,
		"agentID":    agentID,
		"output":     output,
		"logPath":    logPath,
		"logHash":    logHash,
		"success":    success,
		"time":       time.Now().Format(time.RFC3339),
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
