package agent

import (
	"context"
	"iter"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Autumn-27/norma/llm"
)

// Preserve response hints across SDKs that flatten HTTP errors into strings.
// The state is scoped to a single logical call, never shared between profiles.
type responseHintKey struct{}
type responseHint struct {
	mu     sync.Mutex
	status int
	delay  time.Duration
}
type responseFailure struct {
	err    error
	status int
	delay  time.Duration
}

func (e *responseFailure) Error() string             { return e.err.Error() }
func (e *responseFailure) Unwrap() error             { return e.err }
func (e *responseFailure) StatusCode() int           { return e.status }
func (e *responseFailure) RetryAfter() time.Duration { return e.delay }
func (h *responseHint) wrap(err error) error {
	if err == nil {
		return nil
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	return &responseFailure{err: err, status: h.status, delay: h.delay}
}
func captureResponseHint(ctx context.Context, resp *http.Response) {
	h, _ := ctx.Value(responseHintKey{}).(*responseHint)
	if h == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.status = 0
	h.delay = 0
	if resp != nil && resp.StatusCode >= 400 {
		h.status = resp.StatusCode
		h.delay = parseRetryAfter(resp.Header.Get("Retry-After"), time.Now())
	}
}
func parseRetryAfter(raw string, now time.Time) time.Duration {
	raw = strings.TrimSpace(raw)
	if seconds, err := strconv.ParseInt(raw, 10, 64); err == nil && seconds >= 0 && seconds <= (1<<63-1)/int64(time.Second) {
		return time.Duration(seconds) * time.Second
	}
	if t, err := http.ParseTime(raw); err == nil && t.After(now) {
		return t.Sub(now)
	}
	return 0
}

type responseHintProvider struct{ llm.Provider }

func withResponseHints(p llm.Provider) llm.Provider { return &responseHintProvider{p} }
func (p *responseHintProvider) Stream(ctx context.Context, req llm.CompletionRequest) iter.Seq2[llm.StreamEvent, error] {
	return func(yield func(llm.StreamEvent, error) bool) {
		h := &responseHint{}
		ctx = context.WithValue(ctx, responseHintKey{}, h)
		for ev, err := range p.Provider.Stream(ctx, req) {
			if !yield(ev, h.wrap(err)) {
				return
			}
		}
	}
}
func (p *responseHintProvider) Complete(ctx context.Context, req llm.CompletionRequest) (llm.Message, string, llm.Usage, error) {
	h := &responseHint{}
	ctx = context.WithValue(ctx, responseHintKey{}, h)
	msg, stop, usage, err := p.Provider.Complete(ctx, req)
	return msg, stop, usage, h.wrap(err)
}
