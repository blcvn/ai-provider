package helper

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
)

type CryptoHelpers interface {
	Encrypt(text string) (string, error)
	Decrypt(text string) (string, error)
}

type cryptoHelpers struct {
	secretKey []byte
}

func NewCryptoHelpers(secretKey string) CryptoHelpers {
	// Hash the secret key to ensure it's exactly 32 bytes (SHA-256)
	hash := sha256.Sum256([]byte(secretKey))
	return &cryptoHelpers{
		secretKey: hash[:],
	}
}

func (c *cryptoHelpers) Encrypt(text string) (string, error) {
	block, err := aes.NewCipher(c.secretKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(text), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (c *cryptoHelpers) Decrypt(text string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(c.secretKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
