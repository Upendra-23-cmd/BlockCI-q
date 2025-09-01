package blockchain

import (
	"blockci-q/internal/security"
	"crypto/ed25519"
	"encoding/hex"
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
			return fmt.Errorf("prev hash mismatch at index %d", b.Index)
		}
		// index sanity
		if b.Index != i {
			return fmt.Errorf("index mismatch : expected %d got %d ", i, b.Index)
		}

		// Verify signature
		if b.PubKey != "" && b.Signature != "" {
			pubBytes, err := hex.DecodeString(b.PubKey)
			if err != nil {
				return fmt.Errorf("invalid public key at index %d: %w",b.Index,err)
			}
			pubKey := ed25519.PublicKey(pubBytes)

			data , err := b.canonicalData()
			if err != nil {
				return fmt.Errorf("cannot get canonical data at index %d : %w",b.Index,err)
			}
			ok, vErr := security.VerifySignature(pubKey, data, b.Signature)
			if vErr != nil {
				return fmt.Errorf("signature decode error at index %d: %w", b.Index, vErr)
			}
			if !ok {
				return fmt.Errorf("signature verification failed at index %d", b.Index)
			}
		}
	}
	return nil
}