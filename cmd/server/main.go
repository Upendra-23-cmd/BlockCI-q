package main

import (
	"blockci-q/internal/blockchain"
	"blockci-q/internal/core"
	"blockci-q/internal/security"
	"blockci-q/pkg/utils"
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
	ID      string `json:"id"`
	Stage   string `json:"stage"`
	Step    string `json:"step"`
	Cmd     string `json:"cmd"`
	Status  string `json:"status"`
	AgentID string `json:"agentId"`
}

type Server struct {
	mu        sync.Mutex
	ledger    *blockchain.Ledger
	pipelines map[string]*core.Pipeline
	status    map[string]string
	agents    map[string]Agent
	jobs      []Job
	privKey   ed25519.PrivateKey
	pubkey    ed25519.PublicKey
}

func NewServer() *Server {
	ledger, err := blockchain.OpenLedger("./ledger.json")
	if err != nil {
		fmt.Printf("WARN: cannot open ledger: %v\n", err)
	}

	pub, priv, err := ensureServerKey("./keys/server.pub", "./keys/server.priv")
	if err != nil {
		panic(fmt.Sprintf("failed to init server keys: %v", err))
	}

	return &Server{
		ledger:    ledger,
		pipelines: make(map[string]*core.Pipeline),
		status:    make(map[string]string),
		agents:    make(map[string]Agent),
		jobs:      make([]Job, 0),
		privKey:   priv,
		pubkey:    pub,
	}
}

//================================ keys ==================================//

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
		fmt.Println("Generated new server keys")
		return pub, priv, nil
	}
	pub, _ := security.LoadPublicKey(pubPath)
	priv, _ := security.LoadPrivateKey(privPath)
	fmt.Println("Loaded existing server keys ")
	return pub, priv, nil
}

//=============================== pipeline handlers ===============================//

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
	s.status[id] = "pending"
	s.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"id":     id,
		"status": "pending",
	})
}

func (s *Server) handleGetPipelineStatus(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/pipeline/"):]
	status, ok := s.status[id]
	if !ok {
		http.Error(w, "pipeline not found", http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"id": id, "status": status})
}

//=============================== ledger handlers ===============================//

func (s *Server) handleVerifyLedger(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DEBUG: /ledger/verify called")

	if s.ledger == nil {
		http.Error(w, "ledger not initialized", http.StatusInternalServerError)
		return
	}

	if err := s.ledger.VerifyChain(); err != nil {
		http.Error(w, "ledger verification failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("ledger verification ok\n"))
}

//=============================== agent handlers ===============================//

func (s *Server) handleRegisterAgent(w http.ResponseWriter, r *http.Request) {
	var agent Agent
	if err := json.NewDecoder(r.Body).Decode(&agent); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.agents[agent.ID] = agent
	s.mu.Unlock()

	fmt.Println("Agent registered:", agent.ID)
	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(agent)
}

func (s *Server) handleNextJob(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "invalid request path", http.StatusBadRequest)
		return
	}
	agentID := parts[2]

	s.mu.Lock()
	defer s.mu.Unlock()

	for i, job := range s.jobs {
		if job.Status == "pending" {
			s.jobs[i].Status = "running"
			s.jobs[i].AgentID = agentID
			fmt.Printf("Sending job %s to agent %s\n", job.ID, agentID)

			w.Header().Set("Content-type", "application/json")
			json.NewEncoder(w).Encode(job)
			return
		}

	}
	w.WriteHeader(http.StatusNoContent)
}

//=============================== job result handlers ===============================//

func (s *Server) handleJobResult(w http.ResponseWriter, r *http.Request) {
	var result map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	fmt.Println("job result received:", result)

	jobID := fmt.Sprintf("%v", result["id"])

	s.mu.Lock()
	for i := range s.jobs {
		if s.jobs[i].ID == jobID {
			s.jobs[i].Status = "done"
		}
	}
	s.mu.Unlock()

	// record in block chainledger
	idx := s.ledger.NextIndex()
	prev := s.ledger.LastHash()

	agent := fmt.Sprintf("%v", result["agentID"])
	output := fmt.Sprintf("%v", result["output"])
	logPath := fmt.Sprintf("%v", result["logPath"])

	var logHash string
	if logPath != "" && logPath != "<nil>" {
		h, err := utils.HashFile(logPath)
		if err != nil {
			fmt.Printf("⚠️ WARN: cannot hash log file %s: %v\n", logPath, err)
			logHash = utils.HashString(output) // fallback
		} else {
			logHash = h
		}
	} else {
		logHash = utils.HashString(output)
	}

	blk, err := blockchain.NewBlock(idx, "JobResult", jobID, "inline-log", logHash, prev, agent)
	if err != nil {
		http.Error(w, "Failed to create block :"+err.Error(), 500)
		return
	}

	if err := s.ledger.AppendBlocks(blk, s.privKey, s.pubkey); err != nil {
		http.Error(w, "failed to append blocks: "+err.Error(), 500)
		return
	}

	resp := map[string]string{
		"status": "recorded",
		"block":  fmt.Sprintf("%d", blk.Index),
		"time":   time.Now().Format(time.RFC3339),
	}
	json.NewEncoder(w).Encode(resp)
}

//=============================== preload jobs ===============================//

func (s *Server) preloadJobs() {
	s.jobs = append(s.jobs, Job{
		ID:     "job-1",
		Stage:  "Build",
		Step:   "Compile",
		Cmd:    "echo Building",
		Status: "pending",
	})
	s.jobs = append(s.jobs, Job{
		ID:     "job-2",
		Stage:  "Test",
		Step:   "Unit Test",
		Cmd:    "echo Running tests",
		Status: "pending",
	})
}

//=============================== main ===============================//

func main() {
	s := NewServer()
	s.preloadJobs()

	http.HandleFunc("/pipelines", s.handleSubmitPipeline)
	http.HandleFunc("/pipeline/", s.handleGetPipelineStatus)
	http.HandleFunc("/ledger/verify", s.handleVerifyLedger) // ✅ fixed route

	http.HandleFunc("/agent/register", s.handleRegisterAgent)
	http.HandleFunc("/agents/", s.handleNextJob)
	http.HandleFunc("/jobs/result", s.handleJobResult)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("BLOCKCI-Q running on port", port)
	http.ListenAndServe(":"+port, nil)
}
