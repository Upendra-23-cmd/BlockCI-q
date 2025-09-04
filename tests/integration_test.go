
package tests

import (
	"blockci-q/internal/blockchain"
	"blockci-q/internal/security"
	"blockci-q/pkg/utils"
	"os"
	"path/filepath"
	"testing"
)

func createTempLog(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "log.txt")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp log: %v", err)
	}
	return path
}

func TestLedgerBlockAppendAndVerify(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "ledger.jsonl")
	ledger, err := blockchain.OpenLedger(tmpFile)
	if err != nil {
		t.Fatalf("failed to open ledger: %v", err)
	}

	// ✅ Generate keypair for signing
	pub, priv, err := security.GenerateKeyPair()
	if err != nil {
		t.Fatalf("failed to generate keypair: %v", err)
	}

	// Step 1: create fake log
	log1 := createTempLog(t, "hello build step")
	h1, _ := utils.HashFile(log1)
	b1, _ := blockchain.NewBlock(0, "Build", "echo build", log1, h1, "", "agent-1")

	// ✅ Append with proper keys
	if err := ledger.AppendBlocks(b1, priv, pub); err != nil {
		t.Fatalf("failed to append block1: %v", err)
	}

	// Step 2: verify chain
	if err := ledger.VerifyChain(); err != nil {
		t.Fatalf("ledger verify failed unexpectedly: %v", err)
	}

	// Step 3: tamper with block hash
	ledger.Blocks()[0].LogHash = "fake-hash"

	// Step 4: verify should fail
	if err := ledger.VerifyChain(); err == nil {
		t.Errorf("expected tampering detection, but chain verified")
	} else {
		t.Logf("✅ tampering detected: %v", err)
	}
}
