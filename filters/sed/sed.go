package sed

import (
	"io/ioutil"
	"log"
	"regexp"
	"strings"

	"github.com/zalando/skipper/filters"
)

type bodyType int

const (
	sedRequest bodyType = iota
	sedResponse
)

type sed struct {
	regex   *regexp.Regexp
	replace string
}

/*
Substitutes the contents in the body matching a pattern with a given
replacement string. Think of it as using the Unix 'sed' utility with a typical
substitution command of the form 's/regexp/replacement/g'.

The substitution can be applied to a request or response body.

Name: "sed"
*/
func NewSed() filters.Spec { return &sed{} }

// Returns the name of this filter.
func (spec *sed) Name() string {
	return SedName
}

// Creates a new sed filter with the parameters specified in config.
func (spec *sed) CreateFilter(config []interface{}) (filters.Filter, error) {
	if len(config) != 2 {
		return nil, filters.ErrInvalidFilterParameters
	}

	expr, ok := config[0].(string)
	if !ok {
		return nil, filters.ErrInvalidFilterParameters
	}

	replace, ok := config[1].(string)
	if !ok {
		return nil, filters.ErrInvalidFilterParameters
	}

	regex, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}

	return &sed{regex: regex, replace: replace}, nil
}

// Intentionally left with no implementation.
func (_ *sed) Request(_ filters.FilterContext) {}

// Applies this filter's regex to the response body and replaces what was
// matched with the provided replacement string.
func (f *sed) Response(ctx filters.FilterContext) {

	body, err := ioutil.ReadAll(ctx.Response().Body)
	if err != nil {
		log.Println(err)
		return
	}

	transformed := f.regex.ReplaceAllString(string(body), f.replace)

	ctx.Response().Body = ioutil.NopCloser(strings.NewReader(transformed))
	ctx.Response().ContentLength = int64(len(transformed))
}
