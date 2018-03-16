package auth

import (
	"net/http"

	"github.com/zalando/skipper/filters"
)

const (
	OAuthName = "oauth"
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
		client    *http.Client
		tokeninfo string
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

func NewOAuth(tokeninfo string) filters.Spec {
	oauthClient := &oauthClient{
		client:    &http.Client{},
		tokeninfo: tokeninfo,
	}

	return &oauthSpec{client: oauthClient}
}

// Returns the name of this filter
func (spec *oauthSpec) Name() string {
	return OAuthName
}
