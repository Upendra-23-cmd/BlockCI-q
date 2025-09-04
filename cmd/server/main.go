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
	jobs		  []Job
	privKey       ed25519.PrivateKey
	pubkey 		  ed25519.PublicKey
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
		ledger:        ledger,
		pipelines:     make(map[string]*core.Pipeline),	
		status:        make(map[string]string),
		agents:        make(map[string]Agent),
		jobs:          make([]Job,0),
		privKey:       priv,
		pubkey:        pub,
	}
}


//===============================pipeline handler ============================================//

// Ensure keypair exists or generates a new one  
func ensureServerKey(pubPath, privPath string)(ed25519.PublicKey, ed25519.PrivateKey, error){
	if _, err := os.Stat(pubPath); os.IsNotExist(err){

		//generate
		pub, priv, err := security.GenerateKeyPair()
		if err != nil {
			return nil, nil , err
		}
		if err := os.MkdirAll("./keys",0700); err != nil {
			return nil, nil, err
		}
		if err := security.SaveKeyPair(pub, priv, pubPath, privPath); err != nil {
			return nil, nil, err
		}
		fmt.Println("Generated new server keys")
		return pub, priv, nil
	}
	// load existing one
	pub, _ := security.LoadPublicKey(pubPath)
	priv, _ := security.LoadPrivateKey(privPath)
	fmt.Println("Loaded existing server keys ")
	return pub, priv, nil
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



//====================================agent handler============================================//



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


//============================== job results ===========================================//

// POST /jobs/result
func (s *Server) handleJobResult(w http.ResponseWriter, r *http.Request){
	var result map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&result);err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()



	fmt.Println("job result received: ", result)
	
	// record in blockchain ledger
	idx := s.ledger.NextIndex()
	prev := s.ledger.LastHash()

	jobID  := fmt.Sprintf("%v", result["id"])
	agent  := fmt.Sprintf("%v", result["agentID"])
	output := fmt.Sprintf("%v", result["output"])

	// hash the output for immutability
	logHash := utils.HashString(output)


	blk, err := blockchain.NewBlock(idx, "JobResult", jobID, "inline-log", logHash, prev, agent)
	if err != nil {
		http.Error(w ,"Failed to create block :"+err.Error(), 500)
		return
	}

	if err := s.ledger.AppendBlocks(blk, s.privKey, s.pubkey); err != nil {
		http.Error(w, "failed to append blocks: "+err.Error(),500)
		return
	}

	resp := map[string]string{
		"status": "recorded",
		"block" : fmt.Sprintf("%d",blk.Index),
		"time":   time.Now().Format(time.RFC3339),
	}
	json.NewEncoder(w).Encode(resp)

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

// Entry point of the code


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