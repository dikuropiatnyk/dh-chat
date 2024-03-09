package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"io"
	"math/big"

	"golang.org/x/crypto/hkdf"
)

const (
	// AES cipher block size
	KEY_SIZE = 32
)

func EncryptMessage(message string, key []byte) (string, error) {
	// Convert the message to a byte array
	plaintext := []byte(message)
	// Create a new AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	// Create a new GCM block cipher
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	// Create a nonce
	nonce := make([]byte, gcm.NonceSize())
	// Fill the nonce with random data
	if _, err = rand.Read(nonce); err != nil {
		return "", err
	}
	// Encrypt the message
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	// Convert the encrypted message to a string
	return string(ciphertext), nil
}

func DeriveKey(key *big.Int) ([]byte, error) {
	// Create a new HKDF instance
	hkdf := hkdf.New(sha256.New, key.Bytes(), nil, nil)
	// Create a new byte array to store the derived key
	derivedKey := make([]byte, KEY_SIZE)
	// Derive the key
	if _, err := io.ReadFull(hkdf, derivedKey); err != nil {
		return nil, err
	}
	// Return the derived key
	return derivedKey, nil
}
