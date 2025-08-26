package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

func HashFile(path string )(string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err 
	}
	defer f.Close()


	h := sha256.New()
	if _, err := io.Copy(h,f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}