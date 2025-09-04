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
	ID     string  `json:"id"`
	Host string		`json:"host"`
}

type Job struct {
	ID     string `json:"id"`
	Stage  string `json:"stage"`
	Step   string `json:"step"`
	Cmd	   string `json:"cmd"`
}


func main() {
	serverURL := "http://localhost:8080"
	agentID  := "agent-1"

	if err := registeragent(serverURL, agentID);err != nil {
		fmt.Println("X",err)
		os.Exit(1)
	}

	fmt.Println("Agent started , polling for jobs")
	PollJobs(serverURL,agentID)
}



func registeragent(serverURL, id string)error{

	agent := Agent{ID: id, Host: "local host"}
	data, _ := json.Marshal(agent)

	resp, err := http.Post(serverURL+"/agent/register","application/json", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to register agent : %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK{
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("register failed: %s", string(body))
	}

	fmt.Println("Agent registered :", id)
	return nil
}

func PollJobs(serverURL, id string){

	for{
		resp, err := http.Get(fmt.Sprintf("%s/agents/%s/jobs/next", serverURL, id))
		if err != nil {
			fmt.Println("poll error: ", err)
			time.Sleep(5*time.Second)
			continue
		}
		

		if resp.StatusCode == http.StatusNoContent {
			time.Sleep(3*time.Second)
			continue
		}

		var job Job
		if err := json.NewDecoder(resp.Body).Decode(&job); err != nil {
			fmt.Println("Decode job error", err)
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		fmt.Printf("Received job: %s (%s)\n ",job.ID,job.Cmd)

		
		// Run real  job with runner 
		output, sucess := runJob(job, id)

		// Report Result
		reportresult(serverURL,job,id,output,sucess)
	}
}

// Run job using core.runner
func runJob(job Job, agentID string)(string, bool) {
	runner := core.NewRunner()

	//wrap job as minimal pipeline
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
	err := runner.RunPipeline(pipeline)
	if err != nil {
		return fmt.Sprintf("job %s failde : %v", job.ID,err), false
	}
	return fmt.Sprintf("job %s completed successfully", job.ID), true
}


//Report back to the server
func reportresult(serverURL string, job Job, agentID string, output string, sucess bool){
	result := map[string]interface{}{
		"id" :   job.ID,
		"stage": job.Stage,
		"step":  job.Step,
		"cmd":   job.Cmd,
		"agentID": agentID,
		"output" : output,
		"success" : sucess,
		"time" : time.Now().Format(time.RFC3339),
	}

	data ,_ := json.Marshal(result)
	resp, err := http.Post(serverURL+"/jobs/result","application/json", bytes.NewBuffer(data))
	if err != nil {
		fmt.Println("Failed to report result:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println("Reported result:", string(body))
}


