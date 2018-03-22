package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	tokeninfo = "/oauth2/tokeninfo"
)

func TestNewOAuth(t *testing.T) {
	spec := NewOAuth("https://auth.example.org/tokeninfo")

	if spec.Name() != OAuthName {
		t.Errorf("spec.Name(): got %q", spec.Name())
	}
}

func TestValidate(t *testing.T) {

	oauthHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != tokeninfo {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	}

	oauthServer := httptest.NewServer(http.HandlerFunc(oauthHandler))
}
