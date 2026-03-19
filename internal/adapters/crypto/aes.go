// Package crypto provides AES-256-GCM encryption for user secrets (LLM API keys).
//
// The master key is a 32-byte value derived from the ENCRYPTION_KEY environment
// variable (64 hex chars). The key never enters the database — it is held in
// process memory only. Ciphertext is stored as base64(nonce + ciphertext + tag).
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
)

// ErrNoEncryptionKey is returned when ENCRYPTION_KEY is not set or invalid.
var ErrNoEncryptionKey = errors.New("ENCRYPTION_KEY not set or invalid (need 64 hex chars)")

// KeyEncryptor encrypts and decrypts short secrets using AES-256-GCM.
type KeyEncryptor struct {
	gcm cipher.AEAD
}

// NewKeyEncryptor creates a KeyEncryptor from the ENCRYPTION_KEY env var.
// Returns ErrNoEncryptionKey if the env var is missing or not 64 hex chars.
func NewKeyEncryptor() (*KeyEncryptor, error) {
	hexKey := os.Getenv("ENCRYPTION_KEY")
	if len(hexKey) != 64 {
		return nil, ErrNoEncryptionKey
	}
	masterKey, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, ErrNoEncryptionKey
	}
	return NewKeyEncryptorFromBytes(masterKey)
}

// NewKeyEncryptorFromBytes creates a KeyEncryptor from a raw 32-byte key.
func NewKeyEncryptorFromBytes(key []byte) (*KeyEncryptor, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("encryption key must be 32 bytes, got %d", len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("aes.NewCipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("cipher.NewGCM: %w", err)
	}
	return &KeyEncryptor{gcm: gcm}, nil
}

// Encrypt returns base64(nonce + ciphertext + tag) for the given plaintext.
// Returns an empty string for empty plaintext (no key configured).
func (e *KeyEncryptor) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	nonce := make([]byte, e.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}
	ciphertext := e.gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt reverses Encrypt. Returns an empty string for empty ciphertext.
func (e *KeyEncryptor) Decrypt(encoded string) (string, error) {
	if encoded == "" {
		return "", nil
	}
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("base64 decode: %w", err)
	}
	nonceSize := e.gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := e.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("gcm decrypt: %w", err)
	}
	return string(plaintext), nil
}
