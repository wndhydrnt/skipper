/*
Package source implements a custom predicate to match routes
based on the Query Params in URL

It supports checking existence of query params and also checking whether
query params value match to a given regular exp

Examples:

    // Checking existence of a query param
    // matches http://example.org?bb=a&query=withvalue
    example1: QueryParam("query") -> "http://example.org";

    // Even a query param without a value
    // matches http://example.org?bb=a&query=
    example1: QueryParam("query") -> "http://example.org";

    // matches with regexp
    // matches http://example.org?bb=a&query=example
    example1: QueryParam("query", "^example$") -> "http://example.org";

    // matches with regexp and multiple values of query param
    // matches http://example.org?bb=a&query=testing&query=example
    example1: QueryParam("query", "^example$") -> "http://example.org";

*/
package query

import (
	"net/http"
	"regexp"

	"github.com/zalando/skipper/eskip/args"
	"github.com/zalando/skipper/routing"
)

type matchType int

const (
	exists matchType = iota + 1
	matches
)

type predicate struct {
	typ       matchType
	paramName string
	valueExp  *regexp.Regexp
}
type spec struct{}

const name = "QueryParam"

// New creates a new QueryParam predicate specification.
func New() routing.PredicateSpec { return &spec{} }

func (s *spec) Name() string {
	return name
}

func (s *spec) Create(a []interface{}) (routing.Predicate, error) {
	var name, value string
	if err := args.Capture(&name, args.Optional(&value), a); err != nil {
		return nil, err
	}

	typ := exists
	var valueExp *regexp.Regexp
	if value != "" {
		typ = matches
		var err error
		if valueExp, err = regexp.Compile(value); err != nil {
			return nil, err
		}
	}

	return &predicate{typ, name, valueExp}, nil
}

func (p *predicate) Match(r *http.Request) bool {
	queryMap := r.URL.Query()
	vals, ok := queryMap[p.paramName]

	switch p.typ {
	case exists:
		return ok
	case matches:
		if !ok {
			return false
		} else {
			for _, v := range vals {
				if p.valueExp.MatchString(v) {
					return true
				}
			}
			return false
		}

	}

	return false
}
