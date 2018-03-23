package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	tokeninfo   = "/oauth2/tokeninfo"
	tokenPrefix = "Bearer "
)

var (
	testTokens = map[string]string{
		"valid":   "42",
		"timeout": "13",
		"badJson": "666",
	}

	response = map[string][]byte{
		"ok": `{
                  "access_token": "42",
                  "application.read": true,
                  "client_id": "test",
                  "expires_in": 3587,
                  "grant_type": "password",
                  "realm": "/services",
                  "scope": [
                              "application.read",
                              "uid"
                           ],
                  "tpoken_type": "Bearer",
                  "uid": "test-client"
               }`,
		"badJson": `{{}`,
	}
)

func TestNewOAuth(t *testing.T) {
	spec := NewOAuth("https://auth.example.org/tokeninfo")

	if spec.Name() != OAuthName {
		t.Errorf("spec.Name(): got %q", spec.Name())
	}
}

func TestValidate(t *testing.T) {

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != tokeninfo {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		h := r.Header.Get(authHeader)
		if !h.HasPrefix(tokenPrefix) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		switch token := h[len(tokenPrefix):]; token {
		case tokens["badJson"]:
			w.Write(response["badJson"])
		case tokens["timeout"]:
			time.Sleep(2 * time.Second)
			fallthrough
		case tokens["valid"]:
			w.Write(response["ok"])
		default:
			w.WriteHeader(http.StatusUnauthorized)
		}

		return
	}

	oauthServer := httptest.NewServer(http.HandlerFunc(handler))
}
