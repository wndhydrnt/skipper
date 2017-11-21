/*
Package ratelimit provides filters to control the rate limitter settings on the route level.

For detailed documentation of the ratelimit, see https://godoc.org/github.com/zalando/skipper/ratelimit.
*/
package ratelimit

import (
	"time"

	"github.com/zalando/skipper/eskip/args"
	"github.com/zalando/skipper/filters"
	"github.com/zalando/skipper/ratelimit"
)

// RouteSettingsKey is used as key in the context state bag
const RouteSettingsKey = "#ratelimitsettings"

type spec struct {
	typ        ratelimit.Type
	filterName string
}

type filter struct {
	settings ratelimit.Settings
}

// NewLocalRatelimit creates a local measured rate limiting, that is
// only aware of itself. If you have 5 instances with 20 req/s, then
// it would allow 100 req/s to the backend from the same user. A third
// argument can be used to set which part of the request should be
// used to find the same user. Third argument defaults to
// XForwardedForLookuper, meaning X-Forwarded-For Header.
//
// Example:
//
//    backendHealthcheck: Path("/healthcheck")
//    -> localRatelimit(20, "1m")
//    -> "https://foo.backend.net";
//
// Example rate limit per Authorization Header:
//
//    login: Path("/login")
//    -> localRatelimit(3, "1m", "auth")
//    -> "https://login.backend.net";
func NewLocalRatelimit() filters.Spec {
	return &spec{typ: ratelimit.LocalRatelimit, filterName: ratelimit.LocalRatelimitName}
}

// NewRatelimit creates a service rate limiting, that is
// only aware of itself. If you have 5 instances with 20 req/s, then
// it would at max allow 100 req/s to the backend.
//
// Example:
//
//    backendHealthcheck: Path("/healthcheck")
//    -> ratelimit(20, "1s")
//    -> "https://foo.backend.net";
func NewRatelimit() filters.Spec {
	return &spec{typ: ratelimit.ServiceRatelimit, filterName: ratelimit.ServiceRatelimitName}
}

// NewDisableRatelimit disables rate limiting
//
// Example:
//
//    backendHealthcheck: Path("/healthcheck")
//    -> disableRatelimit()
//    -> "https://foo.backend.net";
func NewDisableRatelimit() filters.Spec {
	return &spec{typ: ratelimit.DisableRatelimit, filterName: ratelimit.DisableRatelimitName}
}

func (s *spec) Name() string {
	return s.filterName
}

func serviceRatelimitFilter(a []interface{}) (filters.Filter, error) {
	f := filter{
		settings: ratelimit.Settings{
			Type:     ratelimit.ServiceRatelimit,
			Lookuper: ratelimit.NewSameBucketLookuper(),
		},
	}

	if err := args.Capture(
		&f.settings.MaxHits,
		args.Duration(&f.settings.TimeWindow, time.Second),
		a,
	); err != nil {
		return nil, err
	}

	return &f, nil
}

func localRatelimitFilter(a []interface{}) (filters.Filter, error) {
	var (
		maxHits    int
		timeWindow time.Duration
		lookupType string
		lookuper   ratelimit.Lookuper
	)

	if err := args.Capture(
		&maxHits,
		args.Duration(&timeWindow, time.Second),
		args.Optional(args.Enum(&lookupType, "auth", "ip")),
		a,
	); err != nil {
		return nil, err
	}

	switch lookupType {
	case "auth":
		lookuper = ratelimit.NewAuthLookuper()
	default:
		lookuper = ratelimit.NewXForwardedForLookuper()
	}

	return &filter{
		settings: ratelimit.Settings{
			Type:          ratelimit.LocalRatelimit,
			MaxHits:       maxHits,
			TimeWindow:    timeWindow,
			CleanInterval: 10 * timeWindow,
			Lookuper:      lookuper,
		},
	}, nil
}

func disableFilter(args []interface{}) (filters.Filter, error) {
	return &filter{
		settings: ratelimit.Settings{
			Type: ratelimit.DisableRatelimit,
		},
	}, nil
}

func (s *spec) CreateFilter(args []interface{}) (filters.Filter, error) {
	switch s.typ {
	case ratelimit.ServiceRatelimit:
		return serviceRatelimitFilter(args)
	case ratelimit.LocalRatelimit:
		return localRatelimitFilter(args)
	default:
		return disableFilter(args)
	}
}

// Request stores the configured ratelimit.Settings in the state bag,
// such that it can be used in the proxy to check ratelimit.
func (f *filter) Request(ctx filters.FilterContext) {
	ctx.StateBag()[RouteSettingsKey] = f.settings
}

func (f *filter) Response(filters.FilterContext) {}
