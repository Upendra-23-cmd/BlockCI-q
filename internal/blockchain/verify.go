package blockchain

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
)

// VerifyChain validates all blocks: recompute hash, prev linkage, index sanity and signature.
func (l *Ledger) VerifyChain() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	for i, blk := range l.Blocks {
		// index sanity
		if blk.Index != i {
			return fmt.Errorf("index mismatch at %d: block.Index=%d", i, blk.Index)
		}

		// recompute hash and compare
		expected, err := blk.ComputeHash()
		if err != nil {
			return fmt.Errorf("cannot compute hash for index %d: %w", blk.Index, err)
		}
		if blk.Hash != expected {
			return fmt.Errorf("hash mismatch at index %d", blk.Index)
		}

		// prevHash linkage
		if i > 0 {
			if blk.PrevHash != l.Blocks[i-1].Hash {
				return fmt.Errorf("prevHash mismatch at index %d", blk.Index)
			}
		} else {
			// first block should have empty prevHash (or some genesis rule)
			// no-op, allow empty prevHash
		}

		// require signature and pubKey
		if blk.Signature == "" || blk.PubKey == "" {
			return fmt.Errorf("missing signature or pubKey at index %d", blk.Index)
		}

		// decode pubkey and signature
		pubBytes, err := hex.DecodeString(blk.PubKey)
		if err != nil {
			return fmt.Errorf("invalid pubKey encoding at index %d: %w", blk.Index, err)
		}
		sigBytes, err := hex.DecodeString(blk.Signature)
		if err != nil {
			return fmt.Errorf("invalid signature encoding at index %d: %w", blk.Index, err)
		}

		// verify signature over the block hash
		if !ed25519.Verify(ed25519.PublicKey(pubBytes), []byte(blk.Hash), sigBytes) {
			return fmt.Errorf("signature verification failed at index %d", blk.Index)
		}
	}

	return nil
}
