package auth

import (
	"net/http"

	auth "github.com/abbot/go-http-auth"
	"github.com/zalando/skipper/eskip/args"
	"github.com/zalando/skipper/filters"
)

const (
	Name                      = "basicAuth"
	ForceBasicAuthHeaderName  = "WWW-Authenticate"
	ForceBasicAuthHeaderValue = "Basic realm="
	DefaultRealmName          = "Basic Realm"
)

type basicSpec struct{}

type basic struct {
	authenticator   *auth.BasicAuth
	realmDefinition string
}

func NewBasicAuth() *basicSpec {
	return &basicSpec{}
}

//We do not touch response at all
func (a *basic) Response(filters.FilterContext) {}

// check basic auth
func (a *basic) Request(ctx filters.FilterContext) {
	username := a.authenticator.CheckAuth(ctx.Request())

	if username == "" {
		header := http.Header{}
		header.Set(ForceBasicAuthHeaderName, a.realmDefinition)

		ctx.Serve(&http.Response{
			StatusCode: http.StatusUnauthorized,
			Header:     header,
		})
	}
}

// Creates out basicAuth Filter
// The first params specifies the used htpasswd file
// The second is optional and defines the realm name
func (spec *basicSpec) CreateFilter(a []interface{}) (filters.Filter, error) {
	var configFile string
	realmName := DefaultRealmName
	if err := args.Capture(&configFile, args.Optional(&realmName), a); err != nil {
		return nil, err
	}

	htpasswd := auth.HtpasswdFileProvider(configFile)
	authenticator := auth.NewBasicAuthenticator(realmName, htpasswd)

	return &basic{
		authenticator:   authenticator,
		realmDefinition: ForceBasicAuthHeaderValue + `"` + realmName + `"`,
	}, nil
}

func (spec *basicSpec) Name() string { return Name }
