/*
Package interval implements custom predicates to match routes
only during some period of time.

Package includes three predicates:
Between, Before and After. All predicates can be created using the date
represented as a string in RFC3339 format (see https://golang.org/pkg/time/#pkg-constants),
int64 or float64 number. float64 number will be converted into int64
number.

Between predicate matches only if current date is inside the specified
range of dates. Between predicate requires two dates to be constructed.
Upper boundary must be after lower boundary. Range includes the lower
boundary, but excludes the upper boundary.

Before predicate matches only if current date is before the specified
date. Only one date is required to construct the predicate.

After predicate matches only if current date is after or equal to
the specified date. Only one date is required to construct the predicate.

Examples:

	example1: Path("/zalando") && Between("2016-01-01T12:00:00+02:00", "2016-02-01T12:00:00+02:00") -> "https://www.zalando.de";
	example2: Path("/zalando") && Between(1451642400, 1454320800) -> "https://www.zalando.de";

	example3: Path("/zalando") && Before("2016-02-01T12:00:00+02:00") -> "https://www.zalando.de";
	example4: Path("/zalando") && Before(1454320800) -> "https://www.zalando.de";

	example3: Path("/zalando") && After("2016-01-01T12:00:00+02:00") -> "https://www.zalando.de";
	example4: Path("/zalando") && After(1451642400) -> "https://www.zalando.de";

*/
package interval

import (
	"net/http"
	"time"

	"github.com/zalando/skipper/eskip/args"
	"github.com/zalando/skipper/routing"
)

type intervalType int

const (
	between intervalType = iota
	before
	after
)

type spec struct {
	typ intervalType
}

type predicate struct {
	typ     intervalType
	begin   time.Time
	end     time.Time
	getTime func() time.Time
}

// Creates Between predicate.
func NewBetween() routing.PredicateSpec { return &spec{between} }

// Creates Before predicate.
func NewBefore() routing.PredicateSpec { return &spec{before} }

// Creates After predicate.
func NewAfter() routing.PredicateSpec { return &spec{after} }

func (s *spec) Name() string {
	switch s.typ {
	case between:
		return "Between"
	case before:
		return "Before"
	case after:
		return "After"
	default:
		panic("invalid interval predicate type")
	}
}

func defaultGetTime() time.Time {
	return time.Now()
}

func (s *spec) Create(a []interface{}) (routing.Predicate, error) {
	switch s.typ {
	case before:
		var t time.Time
		if err := args.Capture(&t, a); err != nil {
			return nil, err
		}

		return &predicate{typ: before, end: t, getTime: defaultGetTime}, nil
	case after:
		var t time.Time
		if err := args.Capture(&t, a); err != nil {
			return nil, err
		}

		return &predicate{typ: after, begin: t, getTime: defaultGetTime}, nil
	default:
		var from, to time.Time
		err := args.Capture(&from, &to, a)
		if err != nil || !from.Before(to) {
			return nil, args.ErrInvalidArgs
		}

		return &predicate{between, from, to, defaultGetTime}, nil
	}
}

func (p *predicate) Match(r *http.Request) bool {
	now := p.getTime()

	switch p.typ {
	case between:
		return (p.begin.Before(now) || p.begin.Equal(now)) && p.end.After(now)
	case before:
		return p.end.After(now)
	case after:
		return p.begin.Before(now) || p.begin.Equal(now)
	default:
		return false
	}
}
