package diagnose

import (
	"context"
	"time"
)

type ContextKey string

const NowContextKey ContextKey = "now"

// if the context has a `now` value (set by a unit test), return it.
// otherwise, return `time.Now()`
func now(ctx context.Context) time.Time {
	if now, ok := ctx.Value(NowContextKey).(time.Time); ok {
		return now
	}
	return time.Now()
}
