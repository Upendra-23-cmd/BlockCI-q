package main

import (
	"blockci-q/internal/blockchain"
	"blockci-q/internal/core"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
)

type Agent struct {
	ID   string `json:"id"`
	Host string `json:"host"`
}

type Job struct {
	ID  string `json:"id"`
	Stage string `json:"stage"`
	Step string `json:"step"`
	Cmd  string  `json:"cmd"`
}



type Server struct{
	mu     		  sync.Mutex
	ledger        *blockchain.Ledger
	pipelines	  map[string]*core.Pipeline
	status		  map[string]string
	agents		  map[string]Agent
	jobs			  []Job
}

func NewServer() *Server {
	ledger, err := blockchain.OpenLedger("./ledger.json")
	if err != nil {
		fmt.Printf("WARN: cannot open ledger: %v\n", err)
	}

	return &Server{
		ledger:        ledger,
		pipelines:     make(map[string]*core.Pipeline),	
		status:        make(map[string]string),
		agents:        make(map[string]Agent),
		jobs:          make([]Job,0),
	}
}

// POST /pipelines -> submit a new pipeline YAML

func (s *Server) handleSubmitPipeline(w http.ResponseWriter, r *http.Request){
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "cannot read body",http.StatusBadRequest)
		return
	}

	pipeline , err := core.ParsePipeline(data)
	if err != nil {
		http.Error(w, "invalid pipline", http.StatusBadRequest)
		return
	}

	//Simple id (could use UUID)
	id := fmt.Sprintf("p-%d", len(s.pipelines)+1)

	s.mu.Lock()
	s.pipelines[id]= pipeline
	s.status[id] = "pending"
	s.mu.Unlock()

	w.Header().Set("Content-Type", "appilication/json")
	json.NewEncoder(w).Encode(map[string]string{
		"id":    id,
		"status": "pending",
	})

}

// GET /pipeline/{id}/status
func (s *Server) handleGetPipelineStatus(w http.ResponseWriter, r *http.Request){
	id := r.URL.Path[len("/pipeline/"):]
	status, ok := s.status[id]
	if !ok {
		http.Error(w, "pipeline not found", http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"id": id, "status": status})
}

// GET /ledger/verify -> run Verifychain 
func (s *Server) handleVerifyLedger(w http.ResponseWriter, r *http.Request){
	if err := s.ledger.VerifyChain(); err != nil {
		http.Error(w, "ledger Verification failed :"+err.Error(),http.StatusInternalServerError)
		return
	}
	w.Write([]byte("ledger verification ok"))
}

//	POST /agent/register
func (s *Server) handleRegisterAgent(w http.ResponseWriter, r *http.Request){
	var agent Agent
	if err := json.NewDecoder(r.Body).Decode(&agent); err!=nil{
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.agents[agent.ID]=agent
	s.mu.Unlock()

	fmt.Println("Agent registered:",agent.ID)
	w.Header().Set("Content-type","application/json")
	json.NewEncoder(w).Encode(agent)
}

// GET /agent/{id}/jobs/next
func (s *Server) handleNextJob(w http.ResponseWriter, r *http.Request){
	parts := strings.Split(r.URL.Path,"/")
	if len(parts) < 4 {
		http.Error(w, "invaild request path ", http.StatusBadRequest)
		return
	}
	agentID := parts[2]

	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.jobs) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	job := s.jobs[0]
	s.jobs = s.jobs[1:] 
	fmt.Printf("Sending jobs %s to agent %s\n", job.ID, agentID)

	w.Header().Set("Content-type","application/json")
	json.NewEncoder(w).Encode(job)

}

// POST /jobs/result
func (s *Server) handleJobResult(w http.ResponseWriter, r *http.Request){
	var result map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&result);err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	fmt.Println("job result received: ", result)
	w.WriteHeader(http.StatusOK)
}



// preload fake jobs 
func (s *Server) preloadJobs(){
	s.jobs = append(s.jobs, Job{
		ID: "job-1",
		Stage: "Build",
		Step: "Compile",
		Cmd: "echo Building",
	})
	s.jobs = append(s.jobs, Job{
		ID: "job-2",
		Stage: "Build",
		Step: "Compile",
		Cmd: "echo Running tests",
	})
}




func main() {
	s := NewServer()
	s.preloadJobs()

	http.HandleFunc("/pipelines", s.handleSubmitPipeline)
	http.HandleFunc("/ledger/verify/", s.handleVerifyLedger)
	http.HandleFunc("/pipelines/", s.handleGetPipelineStatus)

	http.HandleFunc("/agent/register", s.handleRegisterAgent)
	http.HandleFunc("/jobs/result", s.handleJobResult)
	http.HandleFunc("/agents/", s.handleNextJob)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("BLOCKCI-Q running on port",port)
	http.ListenAndServe(":"+port,nil)
}