package blockchain

import (
	"blockci-q/internal/security"
	"bufio"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
) 

type Ledger struct {
	Path string
	mu   sync.Mutex
	blocks []*Block
}

// Opensledger  loads an existing ledger.jsonl (if present) or creates an empty one
func OpenLedger(path string )(*Ledger, error) {
	l := &Ledger{Path: path}
	if err := l.load(); err != nil {
		return nil, err
	}
	return l ,nil
}

func (l *Ledger) load() error {
	l.mu.Lock()
	defer l.mu.Unlock()
     
	l.blocks = nil 

	file, err := os.Open(l.Path)
	if errors.Is(err, os.ErrNotExist){
		// Creates an empty file so appends succed later
		f, createErr := os.OpenFile(l.Path, os.O_CREATE|os.O_WRONLY, 0644)
		if createErr != nil {
			return createErr
		}
		_ = f.Close()
		return nil
	}
	if err != nil {
		return err
	}
	defer file.Close()

	sc := bufio.NewScanner(file)
	for sc.Scan() {
		var b Block 
		if err := json.Unmarshal(sc.Bytes(),&b); err != nil {
			return fmt.Errorf("Ledger parse error : %w", err)
		}
		l.blocks = append(l.blocks, &b)
	}
	return sc.Err()
}


// LastHash returns the hash of the last block (or empty if none)
func (l *Ledger) LastHash() string {
	if len(l.blocks) == 0 {
		return ""
	}
	return l.blocks[len(l.blocks)-1].Hash
}

// NextIndex returns the next block index
func (l*Ledger) NextIndex() int {
	return len(l.blocks)
}

// AppendBlocks append a block to memory and file (append-only)
func (l *Ledger) AppendBlocks(b *Block , privkey ed25519.PrivateKey, pubKey ed25519.PublicKey) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	h, err := b.ComputeHash()
	if err  != nil {
		return fmt.Errorf("cannot recompute hash: %w", err)
	}
	b.Hash =h 

	canon, err := b.canonicalData()
	if err != nil {
		return fmt.Errorf("cannot get canonical data : %w",err)
	}
	b.Signature = security.SignData(privkey, canon)
	b.PubKey = hex.EncodeToString(pubKey)


	//basic link check
	if len(l.blocks) > 0 && b.PrevHash != l.blocks[len(l.blocks)-1].Hash {
		return fmt.Errorf("prevHash mismatch: want %s , got %s ", l.blocks[len(l.blocks)-1].Hash, b.PrevHash)
	}

	// append to file 
	f , err := os.OpenFile(l.Path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	line, err := json.Marshal(b)
	if err != nil {
		return err
	}
	if _, err := f.Write(append(line,'\n')); err!= nil {
		return err
	}
	 l.blocks = append(l.blocks ,b)
	 return nil
}

// Blocks (read-only copy)
func (l *Ledger) Blocks() []*Block {
	l.mu.Lock()
	defer l.mu.Unlock()

	out := make([]*Block, 0, len(l.blocks))

    for _, b:= range l.blocks {
        out = append(out, b) // append pointer, not value
    }
	return out
}

