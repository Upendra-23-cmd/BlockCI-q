package tests

import (
	"blockci-q/internal/blockchain"
	"blockci-q/internal/security"
	"blockci-q/pkg/utils"
	"os"
	"path/filepath"
	"testing"
)

// helper to create a dummy log file for hashing
func createTempLog_one(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "log.txt")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp log: %v", err)
	}
	return path
}

// ✅ Test block creation and hashing
func TestNewBlockAndHash(t *testing.T) {
	logPath := createTempLog(t, "hello blockchain")
	logHash, err := utils.HashFile(logPath)
	if err != nil {
		t.Fatalf("failed to hash log: %v", err)
	}

	block, err := blockchain.NewBlock(0, "Build", "echo test", logPath, logHash, "", "test-agent")
	if err != nil {
		t.Fatalf("failed to create block: %v", err)
	}

	// recompute hash and compare
	h, err := block.ComputeHash()
	if err != nil {
		t.Fatalf("failed to recompute hash: %v", err)
	}
	if h != block.Hash {
		t.Errorf("hash mismatch: got %s, want %s", block.Hash, h)
	}
}

// ✅ Test appending multiple blocks to the ledger with signing
func TestLedgerAppendAndVerify(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "ledger.json")
	ledger, err := blockchain.OpenLedger(tmpFile)
	if err != nil {
		t.Fatalf("failed to open ledger: %v", err)
	}

	// generate keys
	pub, priv, _ := security.GenerateKeyPair()

	// create first log
	log1 := createTempLog(t, "step1 output")
	h1, _ := utils.HashFile(log1)
	b1, _ := blockchain.NewBlock(0, "Build", "go build", log1, h1, "", "agent1")

	if err := ledger.AppendBlocks(b1, priv, pub); err != nil {
		t.Fatalf("failed to append block1: %v", err)
	}

	// create second log
	log2 := createTempLog(t, "step2 output")
	h2, _ := utils.HashFile(log2)
	b2, _ := blockchain.NewBlock(1, "Test", "go test ./...", log2, h2, b1.Hash, "agent1")

	if err := ledger.AppendBlocks(b2, priv, pub); err != nil {
		t.Fatalf("failed to append block2: %v", err)
	}

	// verify chain
	if err := ledger.VerifyChain(); err != nil {
		t.Errorf("chain verification failed: %v", err)
	}
}

// ✅ Test tampering detection
func TestTamperingDetection(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "ledger.jsonl")
	ledger, _ := blockchain.OpenLedger(tmpFile)

	pub, priv, _ := security.GenerateKeyPair()

	// create log
	log := createTempLog(t, "secure log")
	h, _ := utils.HashFile(log)
	b, _ := blockchain.NewBlock(0, "Deploy", "echo deploy", log, h, "", "agentX")

	if err := ledger.AppendBlocks(b, priv, pub); err != nil {
		t.Fatalf("append failed: %v", err)
	}

	// simulate tampering
	ledger.Blocks()[0].LogHash = "fakehash"

	// verify should fail
	if err := ledger.VerifyChain(); err == nil {
		t.Errorf("expected verification failure, got success")
	}
}

// ✅ Test ledger persistence (write → reload → verify)
func TestLedgerPersistence(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "ledger.jsonl")
	ledger, _ := blockchain.OpenLedger(tmpFile)

	pub, priv, _ := security.GenerateKeyPair()

	log := createTempLog(t, "persisted log")
	h, _ := utils.HashFile(log)
	b, _ := blockchain.NewBlock(0, "Build", "go build", log, h, "", "agentY")
	_ = ledger.AppendBlocks(b, priv, pub)

	// reopen ledger
	ledger2, err := blockchain.OpenLedger(tmpFile)
	if err != nil {
		t.Fatalf("failed to reopen ledger: %v", err)
	}

	// should still verify
	if err := ledger2.VerifyChain(); err != nil {
		t.Errorf("reloaded ledger verification failed: %v", err)
	}
}
