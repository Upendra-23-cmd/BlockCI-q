package blockchain

import (
	"fmt"
)

// VerifyChain re-computes each block hash and link to detct tampering
func (l *Ledger) VerifyChain() error {
	l.mu.Lock()
	defer l.mu.Unlock()


	for i := range l.blocks {
		b := l.blocks[i]

		// recompute hash
		h , err := b.ComputeHash()
		if err != nil {
			return fmt.Errorf("compute hash for index %d : %w", b.Index, err)
		}
		if h != b.Hash {
			return fmt.Errorf("hash mismatch at index %d" , b.Index)
		}

		// Check link 
		if i > 0 && b.PrevHash != l.blocks[i-1].Hash {
			return fmt.Errorf("Prev hash mismatch at index %d", b.Index)
		}
		// index sanity
		if b.Index != i {
			return fmt.Errorf("index mismatch : expected %d got %d ", i, b,b.Index)
		}
	}
	return nil
}