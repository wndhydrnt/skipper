package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const testEncryptionKey = "1234567890123456"

func TestDecrypt(t *testing.T) {
	keySource, err := NewStaticKeySource([]string{testEncryptionKey})
	assert.NoError(t, err)

	cipherManager, err := NewKeyRotationCipherManager(keySource)
	assert.NoError(t, err)

	encryptedData, err := cipherManager.Encrypt([]byte("PII"))
	assert.NoError(t, err)

	decryptedData, err := cipherManager.Decrypt(encryptedData)
	assert.NoError(t, err)
	assert.Equal(t, []byte("PII"), decryptedData)

	_, err = cipherManager.Decrypt([]byte("not-encrypted"))
	assert.Error(t, err)

	// decrypt something that is too short to be encrypted value
	_, err = cipherManager.Decrypt([]byte("s"))
	assert.Error(t, err)
}
