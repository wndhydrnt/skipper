package auth

import (
	"bytes"
	"fmt"
	"io/ioutil"
)

// KeySource defines an interface for getting Encryption and Decryption keys.
type KeySource interface {
	Keys() (*KeySet, error)
}

// KeySet describes a set of keys which can be used for encryption and
// decryption respectively.
type KeySet struct {
	EncryptionKey  []byte
	DecryptionKeys [][]byte
}

// KeyFilepathSource is a key source which reads keys from a file.
// It assumes one key per line and if there are multiple lines it will treat
// the first line as encryption key and use all lines as decryption keys. This
// allows for key rotation.
type KeyFilepathSource struct {
	FilePath string
}

// Keys return the KeySet after reading from file.
func (s *KeyFilepathSource) Keys() (*KeySet, error) {
	d, err := ioutil.ReadFile(s.FilePath)
	if err != nil {
		return nil, err
	}

	lines := bytes.Split(d, []byte("\n"))
	keys := make([][]byte, 0, len(lines))
	for _, line := range lines {
		if len(line) > 0 {
			keys = append(keys, line)
		}
	}

	if len(keys) == 0 {
		return nil, fmt.Errorf("no keys found")
	}

	return &KeySet{
		EncryptionKey:  keys[0],
		DecryptionKeys: keys,
	}, nil
}

// StaticKeySource is a static list of Keys.
type StaticKeySource struct {
	keys [][]byte
}

// NewStaticKeySource initializes a StaticKeySource.
func NewStaticKeySource(keys []string) (*StaticKeySource, error) {
	if len(keys) == 0 {
		return nil, fmt.Errorf("no keys defined")
	}
	byteKeys := make([][]byte, 0, len(keys))
	for _, key := range keys {
		byteKeys = append(byteKeys, []byte(key))
	}
	return &StaticKeySource{keys: byteKeys}, nil
}

// Keys returns the static set of keys for the source.
func (s *StaticKeySource) Keys() (*KeySet, error) {
	if len(s.keys) == 0 {
		return nil, fmt.Errorf("no keys defined")
	}

	return &KeySet{
		EncryptionKey:  s.keys[0],
		DecryptionKeys: s.keys,
	}, nil
}
