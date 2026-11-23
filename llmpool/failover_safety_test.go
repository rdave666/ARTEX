package llmpool

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Autumn-27/norma/llm"
)

type hintedFailure struct {
	status int
	delay  time.Duration
}

func (e hintedFailure) Error() string             { return "provider unavailable" }
func (e hintedFailure) StatusCode() int           { return e.status }
func (e hintedFailure) RetryAfter() time.Duration { return e.delay }

func TestAllCoolingProfilesAreNotProbed(t *testing.T) {
	for _, stream := range []bool{false, true} {
		t.Run(fmt.Sprint(stream), func(t *testing.T) {
			a, b := okProv("a", "a"), okProv("b", "b")
			reg := NewRegistry(nil, nil)
			reg.Trip(1, "402", true)
			reg.Trip(2, "429", true)
			p := New([]*Member{member(1, "a", 10, a), member(2, "b", 5, b)}, reg)
			var err error
			if stream {
				_, err = drain(p.Stream(t.Context(), llm.CompletionRequest{}))
			} else {
				_, _, _, err = p.Complete(t.Context(), llm.CompletionRequest{})
			}
			if !errors.Is(err, ErrExhausted) {
				t.Errorf("err=%v", err)
			}
			if a.calls+b.calls != 0 {
				t.Errorf("called cooling profiles: %d/%d", a.calls, b.calls)
			}
		})
	}
}
func TestAllOversizedProfilesAreNotProbed(t *testing.T) {
	a := okProv("a", "a")
	m := member(1, "a", 1, a)
	m.WindowTokens = 1
	p := New([]*Member{m}, nil)
	_, err := drain(p.Stream(t.Context(), llm.CompletionRequest{System: []string{"a long system prompt"}}))
	if !errors.Is(err, ErrExhausted) || a.calls != 0 {
		t.Fatalf("err=%v calls=%d", err, a.calls)
	}
}
func TestFailoverDiscardsFailedMetadataPrelude(t *testing.T) {
	a := &fakeProv{events: [][]llm.StreamEvent{{{Type: llm.SEMessageStart, Usage: llm.Usage{InputTokens: 100}}, {Type: llm.SETextDelta}}}, errs: []error{fmt.Errorf("status 429: wait")}}
	b := okProv("b", "backup")
	p := New([]*Member{member(1, "a", 10, a), member(2, "b", 5, b)}, nil)
	acc := llm.NewAccumulator()
	for ev, err := range p.Stream(t.Context(), llm.CompletionRequest{}) {
		if err != nil {
			t.Fatal(err)
		}
		acc.Add(ev)
	}
	if acc.Message().Text() != "backup" || acc.Usage.InputTokens != 0 || b.calls != 1 {
		t.Fatalf("text=%q input=%d backupcalls=%d", acc.Message().Text(), acc.Usage.InputTokens, b.calls)
	}
}
func TestPreludeIsPreservedOnSuccessfulStream(t *testing.T) {
	a := &fakeProv{events: [][]llm.StreamEvent{{{Type: llm.SEMessageStart, Usage: llm.Usage{InputTokens: 17}}, {Type: llm.SETextDelta, Text: "ok"}}}}
	p := New([]*Member{member(1, "a", 1, a)}, nil)
	acc := llm.NewAccumulator()
	for ev, err := range p.Stream(t.Context(), llm.CompletionRequest{}) {
		if err != nil {
			t.Fatal(err)
		}
		acc.Add(ev)
	}
	if acc.Usage.InputTokens != 17 || acc.Message().Text() != "ok" {
		t.Fatalf("usage/text lost: %+v %q", acc.Usage, acc.Message().Text())
	}
}
func TestProviderLocalTimeoutFailsOverWhileParentLives(t *testing.T) {
	a := &fakeProv{errs: []error{fmt.Errorf("request header timeout: %w", context.DeadlineExceeded)}}
	b := okProv("b", "ok")
	p := New([]*Member{member(1, "a", 10, a), member(2, "b", 5, b)}, nil)
	got, err := drain(p.Stream(t.Context(), llm.CompletionRequest{}))
	if err != nil || got != "ok" || b.calls != 1 {
		t.Fatalf("text=%q err=%v backup=%d", got, err, b.calls)
	}
}
func TestTypedRetryHintControlsCooldown(t *testing.T) {
	a := &fakeProv{errs: []error{fmt.Errorf("wrapped: %w", hintedFailure{status: 429, delay: 3 * time.Minute})}}
	b := okProv("b", "ok")
	reg := NewRegistry(nil, nil)
	p := New([]*Member{member(1, "a", 10, a), member(2, "b", 5, b)}, reg)
	_, err := drain(p.Stream(t.Context(), llm.CompletionRequest{}))
	if err != nil {
		t.Fatal(err)
	}
	if statusOf(a.errs[0]) != 429 {
		t.Error("typed status ignored")
	}
	if d := time.Until(reg.Get(1).OpenUntil); d < 179*time.Second {
		t.Fatalf("Retry-After ignored: %v", d)
	}
	_, _ = drain(p.Stream(t.Context(), llm.CompletionRequest{}))
	if a.calls != 1 {
		t.Fatal("rate-limited provider retried during server cooldown")
	}
}
func TestNoFailoverAfterToolUseOrThinking(t *testing.T) {
	for _, ev := range []llm.StreamEvent{{Type: llm.SEToolUseStart, ToolName: "Bash", ToolID: "x"}, {Type: llm.SEThinkingDelta, Text: "reasoning"}} {
		a := &fakeProv{events: [][]llm.StreamEvent{{ev}}, errs: []error{fmt.Errorf("status 503: late error")}}
		b := okProv("b", "no")
		p := New([]*Member{member(1, "a", 10, a), member(2, "b", 5, b)}, nil)
		_, err := drain(p.Stream(t.Context(), llm.CompletionRequest{}))
		if err == nil || b.calls != 0 {
			t.Errorf("unsafe replay after %+v: err=%v backup=%d", ev, err, b.calls)
		}
	}
}

func TestRetryCooldownCannotBeShortenedByConcurrentCompletion(t *testing.T) {
	var saved State
	reg := NewRegistry(func(_ int64, st State) { saved = st }, nil)
	reg.SetPolicy(1, time.Second)
	reg.TripWithDelay(1, "retry later", false, 5*time.Minute)
	deadline := reg.Get(1).OpenUntil
	reg.Pass(1)
	if reg.Get(1).OpenUntil != deadline {
		t.Fatal("in-flight success cleared cooldown")
	}
	reg.TripWithDelay(1, "another failure", false, time.Second)
	if reg.Get(1).OpenUntil != deadline {
		t.Fatal("a shorter retry hint reduced cooldown")
	}
	if saved.OpenUntil != deadline {
		t.Fatal("provider cooldown not persisted")
	}
	restored := NewRegistry(nil, nil)
	restored.Restore(1, saved)
	if !restored.IsOpen(1) {
		t.Fatal("cooldown not retained after restore")
	}
}

func TestCancelledParentDoesNotRotateProviderTimeout(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if shouldFailover(ctx, fmt.Errorf("timeout: %w", context.DeadlineExceeded)) {
		t.Fatal("parent cancellation must not rotate")
	}
}

func TestMetadataOnlyPreludeIsBounded(t *testing.T) {
	a := &fakeProv{events: [][]llm.StreamEvent{make([]llm.StreamEvent, 65)}}
	for i := range a.events[0] {
		a.events[0][i].Type = llm.SEMessageStart
	}
	b := okProv("b", "ok")
	p := New([]*Member{member(1, "a", 10, a), member(2, "b", 5, b)}, nil)
	got, err := drain(p.Stream(t.Context(), llm.CompletionRequest{}))
	if err != nil || got != "ok" || b.calls != 1 {
		t.Fatalf("unbounded or failed prelude: text=%q err=%v backup=%d", got, err, b.calls)
	}
}
