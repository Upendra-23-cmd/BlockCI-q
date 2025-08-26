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
	Timestamp string `json:"timestamp"` // RFC3339
	Stage     string `json:"stage"`
	Step      string `json:"step"`
	LogPath   string `json:"logPath"`
	LogHash   string `json:"logHash"`  // sha256 of the log file content
	PrevHash  string `json:"prevHash"` // hash of previous block
	Hash      string `json:"hash"`     // sha256 over the block's canconial data
	AgentID   string `json:"agentId,omitempty"`
}

// canconical Data return the JSON used to compute the blockchain hash(excluding hash field)
func (b *Block) canonicalData() ([]byte, error) {

	type hashView struct {
		Index     int    `json:"index"`
		Timestamp string `json:"timestamp"`
		Stage     string `json:"stage"`
		Step      string `json:"step"`
		LogPath   string `json:"logPath"`
		LogHash   string `json:"logHash"`
		PrevHash  string `json:"Prevhash"`
		AgentID   string `json:"agentId,omitempty"`
	}
	return json.Marshal(hashView{
		Index:     b.Index,
		Timestamp: b.Timestamp,
		Stage:     b.Stage,
		Step:      b.Step,
		LogPath:   b.LogPath,
		LogHash:   b.LogHash,
		PrevHash:  b.PrevHash,
		AgentID:   b.AgentID,
	})
}

// Compute hash calculates SHA256 over canconical data
func (b *Block) ComputeHash() (string, error) {
	data, err := b.canonicalData()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

// NewBlock constructs a block and seals it with a hash
func NewBlock(index int, stage, step, logPath, logHash, prevHash, agentID string)( *Block , error) {
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
		return nil, fmt.Errorf("compute  block hash: %w", err)
	}
	blk.Hash = h
	return blk, nil
}