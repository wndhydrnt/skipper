package builtin

import (
	"net/url"

	log "github.com/sirupsen/logrus"
	"github.com/zalando/skipper/eskip/args"
	"github.com/zalando/skipper/filters"
)

type spec struct{}

type filter bool

// Returns a filter specification whose filter instances are used to override
// the `proxyPreserveHost` behavior for individual routes.
//
// Instances expect one argument, with the possible values: "true" or "false",
// where "true" means to use the Host header from the incoming request, and
// "false" means to use the host from the backend address.
//
// The filter takes no effect in either case if another filter modifies the
// outgoing host header to a value other than the one in the incoming request
// or in the backend address.
func PreserveHost() filters.Spec { return &spec{} }

func (s *spec) Name() string { return PreserveHostName }

func (s *spec) CreateFilter(a []interface{}) (filters.Filter, error) {
	var preserve string
	if err := args.Capture(
		args.Enum(&preserve, "true", "false"),
		a,
	); err != nil {
		return nil, err
	}

	return filter(preserve == "true"), nil
}

func (preserve filter) Response(_ filters.FilterContext) {}

func (preserve filter) Request(ctx filters.FilterContext) {
	u, err := url.Parse(ctx.BackendUrl())
	if err != nil {
		log.Error("failed to parse backend host in preserveHost filter", err)
		return
	}

	if preserve && ctx.OutgoingHost() == u.Host {
		ctx.SetOutgoingHost(ctx.Request().Host)
	} else if !preserve && ctx.OutgoingHost() == ctx.Request().Host {
		ctx.SetOutgoingHost(u.Host)
	}
}
