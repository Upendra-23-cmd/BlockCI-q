package main

import (
	"blockci-q/internal/blockchain"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var ledger *blockchain.Ledger

func mustJSON(v interface{}) []byte {
	data , err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}


//
// --- Handlers ---
//

// Root welcome page


// Trigger pipeline run (placeholder for now)
func handleRunPipeline(w http.ResponseWriter, r *http.Request) {

	// resp := map[string]string{"status": "pipeline triggered (placeholder)"}
	// _ = json.NewEncoder(w).Encode(resp)

	job := map[string]string{
		"stage": "Build",
		"step":  "complile",
		"cmd":   "echo Building on agent ... ",
	}

	resp, err:= http.Post("http://localhost:9090/run", "application/json", bytes.NewBuffer(mustJSON(job)))
	if err != nil {
		http.Error(w, "failed to contact agent"+err.Error(), http.StatusInternalServerError)
		return
	}

	var agentResp map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&agentResp)
	_ = resp.Body.Close()

	json.NewEncoder(w).Encode(agentResp)

}




func handleRoot(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{
		"message": "Welcome to BlockCI-Q API 🚀",
		"version": "v1.0",
		"routes":  "/api/v1/run, /api/v1/ledger/verify, /api/v1/ledger/latest, /api/v1/ledger/block/{id}, /api/v1/logs/{stage}/{step}",
	}
	_ = json.NewEncoder(w).Encode(resp)
}




// Verify full blockchain ledger
func handleVerifyLedger(w http.ResponseWriter, r *http.Request) {
	err := ledger.VerifyChain()
	status := "ok"
	if err != nil {
		status = "failed: " + err.Error()
	}
	resp := map[string]string{"verification": status}
	_ = json.NewEncoder(w).Encode(resp)
}




// Return latest block in ledger
func handleLatestBlock(w http.ResponseWriter, r *http.Request) {
	blocks := ledger.Blocks()
	if len(blocks) == 0 {
		http.Error(w, "ledger empty", 404)
		return
	}
	_ = json.NewEncoder(w).Encode(blocks[len(blocks)-1])
}




// Get block by ID (index)
func handleGetBlock(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "usage: /api/v1/ledger/block/{id}", 400)
		return
	}
	id := parts[4]

	for _, b := range ledger.Blocks() {
		if fmt.Sprintf("%d", b.Index) == id {
			_ = json.NewEncoder(w).Encode(b)
			return
		}
	}
	http.Error(w, "block not found", 404)
}





// Get logs by stage + step
func handleGetLogs(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "usage: /api/v1/logs/{stage}/{step}", 400)
		return
	}
	stage := parts[3]
	step := parts[4]

	// Find matching log files
	files, err := filepath.Glob(fmt.Sprintf("logs/%s_%s_*.log", stage, step))
	if err != nil || len(files) == 0 {
		http.Error(w, "log not found", 404)
		return
	}

	// Return latest log file content
	content, err := os.ReadFile(files[len(files)-1]) // ✅ using os.ReadFile
	if err != nil {
		http.Error(w, "failed to read log", 500)
		return
	}

	resp := map[string]string{
		"stage":   stage,
		"step":    step,
		"logfile": files[len(files)-1],
		"content": string(content),
	}
	_ = json.NewEncoder(w).Encode(resp)
}





func main() {
	// open or create ledger.jsonl
	// var err error
	// ledger, err = blockchain.OpenLedger("ledger.jsonl")
	// if err != nil {
	// 	log.Fatalf("failed to open ledger: %v", err)
	// }

	mux := http.NewServeMux()

	// Register API routes
	mux.HandleFunc("/", handleRoot)
	mux.HandleFunc("/api/v1/run", handleRunPipeline)
	// mux.HandleFunc("/api/v1/run", handleRunPipeline)
	// mux.HandleFunc("/api/v1/ledger/verify", handleVerifyLedger)
	// mux.HandleFunc("/api/v1/ledger/latest", handleLatestBlock)
	// mux.HandleFunc("/api/v1/ledger/block/", handleGetBlock) // expects /api/v1/ledger/block/{id}
	// mux.HandleFunc("/api/v1/logs/", handleGetLogs)          // expects /api/v1/logs/{stage}/{step}

	fmt.Println("🚀 BlockCI-Q Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
