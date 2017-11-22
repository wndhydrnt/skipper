package args_test

import (
	"log"
	"time"

	"github.com/zalando/skipper/eskip/args"
)

func Example() {
	var (
		maxHits    int
		timeWindow time.Duration
		lookupType string
	)

	a := []interface{}{
		240,
		6,
	}

	if err := args.Capture(
		&maxHits,
		args.Duration(&timeWindow, time.Second),
		args.Optional(args.Enum(&lookupType, "auth", "ip")),
		a,
	); err != nil {
		log.Println(err)
	}
}
