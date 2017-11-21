package builtin

import (
	"github.com/zalando/skipper/eskip/args"
	"github.com/zalando/skipper/filters"
)

const (
	requestCopyFilterName  = "requestCopyHeader"
	responseCopyFilterName = "responseCopyHeader"
	// CopyRequestHeader copies a request header to another proxy
	// request header
	CopyRequestHeader direction = iota
	// CopyResponseHeader copies a proxied response header to the
	// response header
	CopyResponseHeader
)

type direction int

type copySpec struct {
	typ        direction
	filterName string
}
type copyFilter struct {
	typ      direction
	src, dst string
}

// NewCopyRequestHeader creates a filter specification whose instances
// copies a specified source Header to a defined destination Header
// from the request to the proxy request.
func NewCopyRequestHeader() filters.Spec {
	return &copySpec{
		typ:        CopyRequestHeader,
		filterName: requestCopyFilterName,
	}
}

// NewCopyResponseHeader creates a filter specification whose instances
// copies a specified source Header to a defined destination Header
// from the backend response to the proxy response.
func NewCopyResponseHeader() filters.Spec {
	return &copySpec{
		typ:        CopyResponseHeader,
		filterName: responseCopyFilterName,
	}
}

func (s *copySpec) Name() string { return s.filterName }

func (s *copySpec) CreateFilter(a []interface{}) (filters.Filter, error) {
	f := copyFilter{typ: s.typ}
	if err := args.Capture(&f.src, &f.dst, a); err != nil {
		return nil, err
	}

	return &f, nil
}

func (f copyFilter) Request(ctx filters.FilterContext) {
	if f.typ != CopyRequestHeader {
		return
	}

	h := ctx.Request().Header
	if s := h.Get(f.src); s != "" {
		h.Add(f.dst, s)
	}
}

func (f copyFilter) Response(ctx filters.FilterContext) {
	if f.typ != CopyResponseHeader {
		return
	}

	h := ctx.Response().Header
	if s := h.Get(f.src); s != "" {
		h.Add(f.dst, s)
	}
}
