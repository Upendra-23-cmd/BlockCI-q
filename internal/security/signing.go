package security

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
)

// Generate key pair creates a new ed25519 key pair (public+private)
func GenerateKeyPair() ( ed25519.PublicKey,ed25519.PrivateKey, error) {
	pub,priv,err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil , err
	}
	return  pub, priv, nil
}

// SaveKeyPair and save them as hex files
func SaveKeyPair(pub ed25519.PublicKey, priv ed25519.PrivateKey, pubpath,privpath string) error {
	if err := os.WriteFile(pubpath, []byte(hex.EncodeToString(pub)),0600); err != nil {
		return  err
	}
	if err := os.WriteFile(privpath,[]byte(hex.EncodeToString(priv)),0600); err!= nil {
		return  err
	}
	return nil
}

//Loadprivatekey loads an Ed25519 private key from a hex-encoded file
func LoadPrivateKey(path string)(ed25519.PrivateKey, error){
	data , err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	KeyBytes, err := hex.DecodeString(string(data))
	if err != nil {
		return nil, err
	}
	if len(KeyBytes) != ed25519.PrivateKeySize {
		return nil, errors.New("invalid private key size")
	}
	return ed25519.PrivateKey(KeyBytes),nil
}
 
// Loadpublic key loads an Ed5519 public key from hex encoded file
func LoadPublicKey(path string) (ed25519.PublicKey, error) {
	data , err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	KeyBytes, err := hex.DecodeString(string(data))
	if err != nil {
		return nil, err
	}
	if len(KeyBytes) != ed25519.PublicKeySize{
		return  nil, errors.New("invalid public key size")
	}
	return ed25519.PublicKey(KeyBytes), nil
}

// Sign Data Arbitary data using a private key
func SignData(priv ed25519.PrivateKey, data []byte) string{
	sig := ed25519.Sign(priv,data)
	return hex.EncodeToString(sig)
}

// VerifySignature verifies signature of data using a public key 
func VerifySignature(pub ed25519.PublicKey, data []byte, sigHex string)(bool, error){
	sig, err := hex.DecodeString(sigHex)
	if err != nil {
		return false, err
	}
	return ed25519.Verify(pub, data, sig),nil
}

// VerifySignature form hex verifies when the public key is hex encoded
func VerifySignatureFromHex(pubHex string,data []byte, sigHex string )(bool, error){
	pubBytes, err := hex.DecodeString(pubHex)
	if err != nil {
		return false, err
	}
	if len(pubBytes) != ed25519.PublicKeySize{
		return  false, errors.New("invalid public key size")
	}
	return VerifySignature(ed25519.PublicKey(pubBytes), data, sigHex)
}