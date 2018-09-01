package auth

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeyFilepathSource(t *testing.T) {
	content := []byte("secret1\nsecret2\n")
	tmpfile, err := ioutil.TempFile("", "key_filepath_source")
	assert.NoError(t, err)

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content)
	assert.NoError(t, err)
	err = tmpfile.Close()
	assert.NoError(t, err)

	keySource := &KeyFilepathSource{
		FilePath: tmpfile.Name(),
	}

	keySet, err := keySource.Keys()
	assert.NoError(t, err)
	assert.Len(t, keySet.DecryptionKeys, 2)

	// test file with no keys in it
	content = []byte("\n")
	tmpfile, err = ioutil.TempFile("", "key_filepath_source_no_keys")
	assert.NoError(t, err)

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content)
	assert.NoError(t, err)
	err = tmpfile.Close()
	assert.NoError(t, err)

	keySource = &KeyFilepathSource{
		FilePath: tmpfile.Name(),
	}

	keySet, err = keySource.Keys()
	assert.Error(t, err)
}

func TestStaticKeySource(t *testing.T) {
	// test that at least one key must be provided
	_, err := NewStaticKeySource([]string{})
	assert.Error(t, err)

	// test successfully creating a StaticKeySource and getting keys from
	// it.
	keySource, err := NewStaticKeySource([]string{"secret"})
	assert.NoError(t, err)

	keySet, err := keySource.Keys()
	assert.NoError(t, err)
	assert.Len(t, keySet.DecryptionKeys, 1)
	assert.Equal(t, []byte("secret"), keySet.EncryptionKey)
}
