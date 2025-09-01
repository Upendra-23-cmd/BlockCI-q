package main

import (
	"blockci-q/internal/blockchain"
	"blockci-q/internal/core"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
)

type Server struct{
	mu     		  sync.Mutex
	ledger        *blockchain.Ledger
	pipelines	  map[string]*core.Pipeline
	status		  map[string]string
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

func main() {
	s := NewServer()

	http.HandleFunc("/pipelines", s.handleSubmitPipeline)
	http.HandleFunc("/ledger/verify/", s.handleVerifyLedger)
	http.HandleFunc("/pipelines/", s.handleGetPipelineStatus)


	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("BLOCKCI-Q running on port",port)
	http.ListenAndServe(":"+port,nil)
}