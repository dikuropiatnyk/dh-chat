package crypt

import (
	"crypto/aes"
	"crypto/cipher"
)

func DecryptMessage(encryptedMessage string, key []byte) (string, error) {
	// Convert the encrypted message to a byte array
	ciphertext := []byte(encryptedMessage)
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
	// Extract the nonce from the encrypted message
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	// Decrypt the message
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	// Convert the decrypted message to a string
	return string(plaintext), nil
}
