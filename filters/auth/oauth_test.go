package auth

import (
	"testing"
)

var ()

func TestNewOAuth(t *testing.T) {
	url := "https://auth.example.org/tokeninfo"
	spec := NewOAuth(url)

	if spec.Name() != OAuthName {
		t.Errorf("NewOAuth.Name(): got %q", spec.Name())
	}
}
