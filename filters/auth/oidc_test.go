package auth

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_getTimestampFromState(t *testing.T) {
	keySource, err := NewStaticKeySource([]string{testEncryptionKey})
	assert.NoError(t, err)

	cipherManager, err := NewKeyRotationCipherManager(keySource)
	assert.NoError(t, err)

	nonce, err := cipherManager.Nonce()
	assert.NoError(t, err)

	nonceHex := fmt.Sprintf("%x", nonce)
	statePlain := createState(nonceHex)

	ts := getTimestampFromState([]byte(statePlain), len(nonceHex))
	if time.Now().After(ts) {
		t.Errorf("now is after time from state but should be before: %s", ts)
	}
}

func Test_createState(t *testing.T) {
	in := "foo"
	out := createState(in)
	ts := fmt.Sprintf("%d", time.Now().Add(1*time.Minute).Unix())
	if len(out) != len(in)+len(ts)+secretSize {
		t.Errorf("createState returned string size is wrong: %s", out)
	}
}
