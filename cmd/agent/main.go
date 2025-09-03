package main

import (
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

		var job map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&job); err != nil {
			fmt.Println("Decode job error", err)
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		fmt.Println("Received job: ",job)

		// run fake execution 
		result := map[string]interface{}{
			"jobID": job["id"],
			"agent" : id,
			"output": "simulated success",
			"status": "done",
			"time"  : time.Now().Format(time.RFC3339),
		}

		data, _ := json.Marshal(result)
		http.Post(serverURL+"/jobs/result","application/json",bytes.NewBuffer(data))
	}
}


