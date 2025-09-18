package blockchain

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type Ledger struct {
	mu     sync.Mutex
	Blocks []*Block
	path   string
}

// OpenLedger loads an existing ledger file or creates a new in-memory ledger.
// Ledger file format: JSON lines (one JSON block per line).
func OpenLedger(path string) (*Ledger, error) {
	l := &Ledger{
		Blocks: make([]*Block, 0),
		path:   path,
	}

	// If file missing, create empty file
	if _, err := os.Stat(path); os.IsNotExist(err) {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		_ = f.Close()
		return l, nil
	}

	// Read file and decode JSON lines
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return l, nil
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	for dec.More() {
		var blk Block
		if err := dec.Decode(&blk); err != nil {
			return nil, fmt.Errorf("failed to decode ledger entry: %w", err)
		}
		l.Blocks = append(l.Blocks, &blk)
	}
	return l, nil
}

// AppendBlocks appends a block into the ledger, signs it with server's priv key,
// stores hex pubkey, persists to disk (JSONL), and keeps it in memory.
func (l *Ledger) AppendBlocks(b *Block, priv ed25519.PrivateKey, pub ed25519.PublicKey) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// recompute and set hash to be sure block canonical fields match
	h, err := b.ComputeHash()
	if err != nil {
		return fmt.Errorf("cannot recompute block hash: %w", err)
	}
	b.Hash = h

	// prevHash check
	if len(l.Blocks) > 0 {
		last := l.Blocks[len(l.Blocks)-1]
		if b.PrevHash != last.Hash {
			return fmt.Errorf("prevHash mismatch: expected %s, got %s", last.Hash, b.PrevHash)
		}
	}

	// Sign the block hash with server private key and set pubkey
	if len(priv) == 0 {
		return fmt.Errorf("private key is empty, cannot sign block")
	}
	sig := ed25519.Sign(priv, []byte(b.Hash))
	b.Signature = hex.EncodeToString(sig)
	b.PubKey = hex.EncodeToString(pub)

	// Append to file (create if missing)
	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("open ledger file: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	if err := enc.Encode(b); err != nil {
		return fmt.Errorf("write ledger file: %w", err)
	}

	// Push into memory
	l.Blocks = append(l.Blocks, b)
	return nil
}

// NextIndex returns the next block index (not locking for heavy concurrency)
func (l *Ledger) NextIndex() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.Blocks)
}

// LastHash returns the last block hash (or empty if none)
func (l *Ledger) LastHash() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.Blocks) == 0 {
		return ""
	}
	return l.Blocks[len(l.Blocks)-1].Hash
}
