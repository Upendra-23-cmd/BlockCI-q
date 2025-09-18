package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// Block is a tamper-evident record for one pipeline step
type Block struct {
	Index     int    `json:"index"`
	Timestamp string `json:"timestamp"`
	Stage     string `json:"stage"`
	Step      string `json:"step"`
	LogPath   string `json:"logPath"`
	LogHash   string `json:"logHash"`
	PrevHash  string `json:"prevHash"`
	Hash      string `json:"hash"`
	AgentID   string `json:"agentId"`
	Signature string `json:"signature"`
	PubKey    string `json:"pubKey"`
}

// canonicalData returns the JSON bytes used to compute the block hash.
// It intentionally excludes Hash, Signature and PubKey.
func (b *Block) canonicalData() ([]byte, error) {
	// Use a stable view for hashing
	view := struct {
		Index    int    `json:"index"`
		Timestamp string `json:"timestamp"`
		Stage    string `json:"stage"`
		Step     string `json:"step"`
		LogPath  string `json:"logPath"`
		LogHash  string `json:"logHash"`
		PrevHash string `json:"prevHash"`
		AgentID  string `json:"agentId"`
	}{
		Index:     b.Index,
		Timestamp: b.Timestamp,
		Stage:     b.Stage,
		Step:      b.Step,
		LogPath:   b.LogPath,
		LogHash:   b.LogHash,
		PrevHash:  b.PrevHash,
		AgentID:   b.AgentID,
	}
	return json.Marshal(view)
}

// ComputeHash calculates SHA256 over canonicalData
func (b *Block) ComputeHash() (string, error) {
	data, err := b.canonicalData()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

// NewBlock constructs a block and computes its hash (no signature yet)
func NewBlock(index int, stage, step, logPath, logHash, prevHash, agentID string) (*Block, error) {
	blk := &Block{
		Index:     index,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Stage:     stage,
		Step:      step,
		LogPath:   logPath,
		LogHash:   logHash,
		PrevHash:  prevHash,
		AgentID:   agentID,
	}

	h, err := blk.ComputeHash()
	if err != nil {
		return nil, fmt.Errorf("compute block hash: %w", err)
	}
	blk.Hash = h
	return blk, nil
}
