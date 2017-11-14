package builtin

import (
	"github.com/zalando/skipper/args"
	"github.com/zalando/skipper/filters"
)

type statusSpec struct{}

type statusFilter int

// Creates a filter specification whose instances set the
// status of the response to a fixed value regardless of
// backend response.
func NewStatus() filters.Spec { return new(statusSpec) }

func (s *statusSpec) Name() string { return StatusName }

func (s *statusSpec) CreateFilter(a []interface{}) (filters.Filter, error) {
	var value int
	if err := args.Capture(&value, a); err != nil {
		return nil, err
	}

	return statusFilter(value), nil
}

func (f statusFilter) Request(filters.FilterContext) {}

func (f statusFilter) Response(ctx filters.FilterContext) {
	ctx.Response().StatusCode = int(f)
}
