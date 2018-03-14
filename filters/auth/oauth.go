package auth

import (
	"net/http"
)

const (
	Name = "oauth"
)

type (
	oauthInfo struct {
		Realm string
		Scope []string
		Uid   string
	}

	oauthSpec struct {
		client *oauthClient
	}

	oauthClient struct {
		client       *http.Client
		tokeninfoURL string
	}

	oauthFilter struct {
		client oauthClient
		realm  string
		scopes []string
	}
)

const (
	authHeader = "Authorization"
)

// Returns the name of this filter
func (spec *authSpec) Name() string {
	return Name
}
