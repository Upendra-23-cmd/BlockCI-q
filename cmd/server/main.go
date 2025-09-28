package main

import (
	"blockci-q/internal/blockchain"
	"blockci-q/internal/core"
	"blockci-q/internal/security"
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
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

type Server struct {
	mu        sync.Mutex
	ledger    *blockchain.Ledger
	pipelines map[string]*core.Pipeline
	status    map[string]map[string]string // pipelineID -> step -> status
	agents    map[string]Agent
	jobs      []Job
	privKey   ed25519.PrivateKey
	pubKey    ed25519.PublicKey
}

//========================= INIT ===============================//

func NewServer() *Server {
	ledger, err := blockchain.OpenLedger("./ledger.json")
	if err != nil {
		fmt.Printf("⚠️ WARN: cannot open ledger: %v\n", err)
	}

	pub, priv, err := ensureServerKey("./keys/server.pub", "./keys/server.priv")
	if err != nil {
		panic(fmt.Sprintf("❌ failed to init server keys: %v", err))
	}

	return &Server{
		ledger:    ledger,
		pipelines: make(map[string]*core.Pipeline),
		status:    make(map[string]map[string]string),
		agents:    make(map[string]Agent),
		jobs:      make([]Job, 0),
		privKey:   priv,
		pubKey:    pub,
	}
}

func ensureServerKey(pubPath, privPath string) (ed25519.PublicKey, ed25519.PrivateKey, error) {
	if _, err := os.Stat(pubPath); os.IsNotExist(err) {
		pub, priv, err := security.GenerateKeyPair()
		if err != nil {
			return nil, nil, err
		}
		if err := os.MkdirAll("./keys", 0700); err != nil {
			return nil, nil, err
		}
		if err := security.SaveKeyPair(pub, priv, pubPath, privPath); err != nil {
			return nil, nil, err
		}
		fmt.Println("🔑 Generated new server keys")
		return pub, priv, nil
	}
	pub, _ := security.LoadPublicKey(pubPath)
	priv, _ := security.LoadPrivateKey(privPath)
	fmt.Println("🔑 Loaded existing server keys")
	return pub, priv, nil
}

//========================= PIPELINE ===============================//

// POST /pipelines -> submit a new pipeline YAML
func (s *Server) handleSubmitPipeline(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "cannot read body", http.StatusBadRequest)
		return
	}

	pipeline, err := core.ParsePipeline(data)
	if err != nil {
		http.Error(w, "invalid pipeline", http.StatusBadRequest)
		return
	}

	id := fmt.Sprintf("p-%d", len(s.pipelines)+1)

	s.mu.Lock()
	s.pipelines[id] = pipeline
	s.status[id] = make(map[string]string)

	// flatten pipeline → jobs
	jobCount := 0
	for _, stage := range pipeline.Stages {
		for _, step := range stage.Steps {
			jobID := fmt.Sprintf("%s-job-%d", id, jobCount+1)
			job := Job{
				ID:         jobID,
				Stage:      stage.Name,
				Step:       step.Name,
				Cmd:        step.Run,
				PipelineID: id,
			}
			s.jobs = append(s.jobs, job)
			s.status[id][fmt.Sprintf("%s:%s", stage.Name, step.Name)] = "pending"
			jobCount++
		}
	}
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"id":     id,
		"status": "pending",
	})
}

// GET /pipelines/{id}/status
func (s *Server) handleGetPipelineStatus(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/pipelines/")
	s.mu.Lock()
	defer s.mu.Unlock()

	status, ok := s.status[id]
	if !ok {
		http.Error(w, "pipeline not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(status)
}

//========================= LEDGER ===============================//

func (s *Server) handleVerifyLedger(w http.ResponseWriter, r *http.Request) {
	if err := s.ledger.VerifyChain(); err != nil {
		http.Error(w, "ledger verification failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte("✅ ledger verification OK"))
}

//========================= AGENT ===============================//

func (s *Server) handleRegisterAgent(w http.ResponseWriter, r *http.Request) {
	var agent Agent
	if err := json.NewDecoder(r.Body).Decode(&agent); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	s.mu.Lock()
	s.agents[agent.ID] = agent
	s.mu.Unlock()

	fmt.Println("🤝 Agent registered:", agent.ID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(agent)
}

// GET /agents/{id}/jobs/next
func (s *Server) handleNextJob(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "invalid request path", http.StatusBadRequest)
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

	fmt.Printf("📤 Dispatching job %s (%s:%s) from pipeline %s → agent %s\n",
		job.ID, job.Stage, job.Step, job.PipelineID, agentID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(job)
}

//========================= JOB RESULTS ===============================//

func (s *Server) handleJobResult(w http.ResponseWriter, r *http.Request) {
	var result map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	fmt.Printf("📩 Job result received from agent %s for job %s (pipeline %s)\n",
		result["agentID"], result["id"], result["pipelineId"])

	idx := s.ledger.NextIndex()
	prev := s.ledger.LastHash()

	stage := fmt.Sprintf("%v", result["stage"])
	step := fmt.Sprintf("%v", result["step"])
	logPath := fmt.Sprintf("%v", result["logPath"])
	logHash := fmt.Sprintf("%v", result["logHash"])
	agent := fmt.Sprintf("%v", result["agentID"])
	pipelineID := fmt.Sprintf("%v", result["pipelineId"])

	blk, err := blockchain.NewBlock(idx, stage, step, logPath, logHash, prev, agent)
	if err != nil {
		http.Error(w, "failed to create block: "+err.Error(), 500)
		return
	}

	if err := s.ledger.AppendBlocks(blk, s.privKey, s.pubKey); err != nil {
		http.Error(w, "failed to append block: "+err.Error(), 500)
		return
	}

	// update pipeline status
	s.mu.Lock()
	s.status[pipelineID][fmt.Sprintf("%s:%s", stage, step)] = "done"
	s.mu.Unlock()

	resp := map[string]string{
		"status": "recorded",
		"block":  fmt.Sprintf("%d", blk.Index),
		"time":   time.Now().Format(time.RFC3339),
	}
	json.NewEncoder(w).Encode(resp)
}

//========================= BOOTSTRAP ===============================//

func main() {
	s := NewServer()

	http.HandleFunc("/pipelines", s.handleSubmitPipeline)
	http.HandleFunc("/pipelines/", s.handleGetPipelineStatus)
	http.HandleFunc("/ledger/verify", s.handleVerifyLedger)

	http.HandleFunc("/agent/register", s.handleRegisterAgent)
	http.HandleFunc("/agents/", s.handleNextJob)
	http.HandleFunc("/jobs/result", s.handleJobResult)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("🚀 BLOCKCI-Q Server running on port", port)
	http.ListenAndServe(":"+port, nil)
}
