package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
)

type JobRequest struct {
	Stage string `json:"stage"`
	Step  string `json:"step"`
	Cmd   string `json:"cmd"`
}

type JobResponse struct {
	Stage   string `json:"stage"`
	Step    string `json:"step"`
	Success string `json:"sucess"`
	Output  string `json:"output"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/run", handleRunJob)

	fmt.Println("Agent Running on http://localhost:9090 ")
	log.Fatal(http.ListenAndServe(":9090", mux))
}

func handleRunJob(w http.ResponseWriter, r *http.Request) {
	var job JobRequest
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Printf("Agent running job %s - %s\n", job.Stage, job.Step)

	//Execute command
	cmd := exec.Command("sh", "-c", job.Cmd)
	output, err := cmd.CombinedOutput()
	status := "Success"
	if err != nil {
		status = "Error"
	}

	resp := JobResponse{
		Stage:   job.Stage,
		Step:    job.Step,
		Success: status,
		Output:  string(output),
	}

	_ = json.NewEncoder(w).Encode(resp)
}
