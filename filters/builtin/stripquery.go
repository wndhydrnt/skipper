package builtin

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"

	"github.com/zalando/skipper/eskip/args"
	"github.com/zalando/skipper/filters"
)

type stripQuery struct {
	preserveAsHeader bool
}

// Returns a filter Spec to strip query parameters from the request and
// optionally transpose them to request headers.
//
// It always removes the query parameter from the request URL, and if the
// first filter parameter is "true", preserves the query parameter in the form
// of x-query-param-<queryParamName>: <queryParamValue> headers, so that
// ?foo=bar becomes x-query-param-foo: bar
//
// Name: "stripQuery".
func NewStripQuery() filters.Spec { return &stripQuery{} }

// "stripQuery"
func (_ *stripQuery) Name() string { return StripQueryName }

// copied from textproto/reader
func validHeaderFieldByte(b byte) bool {
	return ('A' <= b && b <= 'Z') ||
		('a' <= b && b <= 'z') ||
		('0' <= b && b <= '9') ||
		b == '-'
}

// make sure we don't generate invalid headers
func sanitize(input string) string {
	toAscii := strconv.QuoteToASCII(input)
	var b bytes.Buffer
	for _, i := range toAscii {
		if validHeaderFieldByte(byte(i)) {
			b.WriteRune(i)
		}
	}
	return b.String()
}

// Strips the query parameters and optionally preserves them in the X-Query-Param-xyz headers.
func (f *stripQuery) Request(ctx filters.FilterContext) {
	r := ctx.Request()
	if r == nil {
		return
	}

	url := r.URL
	if url == nil {
		return
	}

	if !f.preserveAsHeader {
		url.RawQuery = ""
		return
	}

	q := url.Query()
	for k, vv := range q {
		for _, v := range vv {
			if r.Header == nil {
				r.Header = http.Header{}
			}
			r.Header.Add(fmt.Sprintf("X-Query-Param-%s", sanitize(k)), v)
		}
	}

	url.RawQuery = ""
}

// Noop.
func (_ *stripQuery) Response(ctx filters.FilterContext) {}

// Creates instances of the stripQuery filter. Accepts one optional parameter:
// "true", in order to preserve the stripped parameters in the request header.
func (_ *stripQuery) CreateFilter(a []interface{}) (filters.Filter, error) {
	var preserveAsHeader string
	if err := args.Capture(
		args.Optional(args.Enum(&preserveAsHeader, "true", "false")),
		a,
	); err != nil {
		return nil, err
	}

	return &stripQuery{preserveAsHeader == "true"}, nil
}
