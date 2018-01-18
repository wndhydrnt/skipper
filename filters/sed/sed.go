// Copyright 2015 Zalando SE
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sed

import (
	"io/ioutil"
	"log"
	"regexp"
	"strings"

	"github.com/zalando/skipper/filters"
)

type sedType int

const (
	sedRequest sedType = iota
	sedResponse

	SedRequestName  = "sedRequest"
	SedResponseName = "sedResponse"
)

/*
Holds the regular expression pattern and replacement string to be applied for
a given filter type.
*/
type sed struct {
	typ     sedType
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
	switch spec.typ {
	case sedRequest:
		return SedRequestName
	case sedResponse:
		return SedResponseName
	default:
		panic("invalid header type")
	}
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

/*
If the filter is defined as 'sedRequest', matches the request body against this
filter's regex. If successful, the portion which was matched is replaced with
the provided replacement string.
*/
func (f *sed) Request(ctx filters.FilterContext) {
	if f.typ == sedRequest {
		applyRegexSub(ctx.Request().Body, f)
	}
}

/*
If the filter is defined as 'sedResponse', matches the response body against
this filter's regex. If successful, the portion which was matched is replaced
with the provided replacement string.
*/
func (f *sed) Response(ctx filters.FilterContext) {
	if f.typ == sedResponse {
		transformed, err = f.applyRegexSub(ctx.Response().Body)
		if err != nil {
			log.Println(err)
			return
		}

	}
}

// Applies the regex
func (f *sed) applyRegexSub(ctx filters.FilterContext) {
	var body []byte
	var err error
	body, err := ioutil.ReadAll(ctx.Response().Body)
	if err != nil {
		log.Println(err)
		return
	}

	transformed := f.regex.ReplaceAllString(string(body), f.replace)

	ctx.Response().Body = ioutil.NopCloser(strings.NewReader(transformed))
	ctx.Response().ContentLength = int64(len(transformed))
}
