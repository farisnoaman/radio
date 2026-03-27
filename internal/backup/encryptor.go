package backup

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"os"
)

// Encryptor provides AES-GCM file encryption for backups
type Encryptor struct{}

// NewEncryptor creates a new Encryptor instance
func NewEncryptor() *Encryptor {
	return &Encryptor{}
}

// EncryptFile encrypts a file using AES-GCM
func (e *Encryptor) EncryptFile(filepath, key string) (string, error) {
	// Read file
	data, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}

	// Create cipher block from key (pad or truncate key to 32 bytes for AES-256)
	keyBytes := []byte(key)
	for len(keyBytes) < 32 {
		keyBytes = append(keyBytes, 0)
	}
	if len(keyBytes) > 32 {
		keyBytes = keyBytes[:32]
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt and authenticate
	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	// Write encrypted file
	encryptedPath := filepath + ".enc"
	if err := os.WriteFile(encryptedPath, ciphertext, 0600); err != nil {
		return "", err
	}

	return encryptedPath, nil
}

// DecryptFile decrypts an encrypted file
func (e *Encryptor) DecryptFile(filepath, key string) (string, error) {
	// Read encrypted file
	data, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}

	// Create cipher block from key (pad or truncate key to 32 bytes for AES-256)
	keyBytes := []byte(key)
	for len(keyBytes) < 32 {
		keyBytes = append(keyBytes, 0)
	}
	if len(keyBytes) > 32 {
		keyBytes = keyBytes[:32]
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Extract nonce
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	// Write decrypted file
	decryptedPath := filepath[:len(filepath)-4] // Remove .enc
	if err := os.WriteFile(decryptedPath, plaintext, 0600); err != nil {
		return "", err
	}

	return decryptedPath, nil
}
