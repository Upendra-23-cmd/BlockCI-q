package main

import (
	"blockci-q/internal/blockchain"
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: blockci <inspect|verify|tamper> <ledger.jsonl> [blockIndex]")
		os.Exit(1)
	}

	cmd := os.Args[1]
	ledgerPath := os.Args[2]

	ledger, err := blockchain.OpenLedger(ledgerPath)
	if err != nil {
		fmt.Printf("Failed to open ledger: %v\n", err)
		os.Exit(1)
	}

	switch cmd {

	case "inspect":
		for _, b := range ledger.Blocks() {
			fmt.Printf("Index=%d Stage=%s Step=%s Hash=%s\n",
				b.Index, b.Stage, b.Step, b.Hash[:16])
		}

	case "verify":
		if err := ledger.VerifyChain(); err != nil {
			fmt.Printf("❌ Verification FAILED: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ Ledger verification OK")

	case "tamper":
		if len(os.Args) < 4 {
			fmt.Println("Usage: blockci tamper <ledger.jsonl> <blockIndex>")
			os.Exit(1)
		}

		// pick block index
		var idx int
		fmt.Sscanf(os.Args[3], "%d", &idx)

		blocks := ledger.Blocks()
		if idx < 0 || idx >= len(blocks) {
			fmt.Printf("Invalid block index %d\n", idx)
			os.Exit(1)
		}

		// Corrupt the logHash
		blocks[idx].LogHash = "FAKE_HASH_TAMPERED"

		// Overwrite file with corrupted blocks
		f, err := os.Create(ledgerPath)
		if err != nil {
			fmt.Printf("Failed to reopen ledger for tampering: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()

		enc := json.NewEncoder(f)
		for _, b := range blocks {
			if err := enc.Encode(b); err != nil {
				fmt.Printf("Failed to rewrite ledger: %v\n", err)
				os.Exit(1)
			}
		}

		fmt.Printf("⚠️ Tampered block %d (LogHash set to FAKE_HASH_TAMPERED)\n", idx)

	default:
		fmt.Println("Unknown command:", cmd)
		os.Exit(1)
	}
}
