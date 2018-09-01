package auth

import (
	"crypto/aes"
	"crypto/cipher"
	crand "crypto/rand"
	"errors"
	"fmt"
	"io"
)

// CipherManager provides Encryption and Decryption.
type CipherManager interface {
	Encrypt(plaintext []byte) ([]byte, error)
	Decrypt(secret []byte) ([]byte, error)
	Nonce() ([]byte, error)
}

// KeyRotationCipherManager is a Cipher Manager which can Encrypt and Decrypt
// while supporting key rotation.
type KeyRotationCipherManager struct {
	encryptionCipher  cipher.AEAD
	decryptionCiphers []cipher.AEAD
}

// NewKeyRotationCipherManager initializes a new KayRotationCipherManager.
func NewKeyRotationCipherManager(keySource KeySource) (*KeyRotationCipherManager, error) {
	keySet, err := keySource.Keys()
	if err != nil {
		return nil, fmt.Errorf("failed to get keys: %v", err)
	}

	//key has to be 16 or 32 byte
	block, err := aes.NewCipher(keySet.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create new cipher: %v", err)
	}

	encryptionAESgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create new GCM: %v", err)
	}

	manager := &KeyRotationCipherManager{
		encryptionCipher: encryptionAESgcm,
	}

	decryptionCiphers := make([]cipher.AEAD, 0, len(keySet.DecryptionKeys))
	for _, key := range keySet.DecryptionKeys {
		//key has to be 16 or 32 byte
		block, err := aes.NewCipher(key)
		if err != nil {
			return nil, fmt.Errorf("failed to create new cipher: %v", err)
		}

		aesgcm, err := cipher.NewGCM(block)
		if err != nil {
			return nil, fmt.Errorf("failed to create new GCM: %v", err)
		}

		decryptionCiphers = append(decryptionCiphers, aesgcm)
	}

	manager.decryptionCiphers = decryptionCiphers
	return manager, nil
}

// Encrypt encrypts plaintext and returns the encrypted bytes.
func (m *KeyRotationCipherManager) Encrypt(plaintext []byte) ([]byte, error) {
	nonce, err := m.Nonce()
	if err != nil {
		return nil, err
	}
	return m.encryptionCipher.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt decrypts cipherText and returns the decrypted bytes.
func (m *KeyRotationCipherManager) Decrypt(cipherText []byte) ([]byte, error) {
	var err error
	for _, aead := range m.decryptionCiphers {
		nonceSize := aead.NonceSize()
		if len(cipherText) < nonceSize {
			err = errors.New("failed to decrypt, ciphertext too short")
			continue
		}
		nonce, input := cipherText[:nonceSize], cipherText[nonceSize:]
		var secret []byte
		secret, err = aead.Open(nil, nonce, input, nil)
		if err == nil {
			return secret, nil
		}
	}

	return nil, err
}

// Nonce returns a new nonce for the internal cipher.
func (m *KeyRotationCipherManager) Nonce() ([]byte, error) {
	nonce := make([]byte, m.encryptionCipher.NonceSize())
	if _, err := io.ReadFull(crand.Reader, nonce); err != nil {
		return nil, err
	}
	return nonce, nil
}
