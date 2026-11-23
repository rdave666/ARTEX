package server

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Autumn-27/artex/agent"
)

func TestWorkStallWatcherTracksProgressAndCancelsIdleRun(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	defer cancel(nil)
	progress := new(atomic.Int64)
	progress.Store(time.Now().UnixNano())
	stop := startWorkStallWatcher(ctx, 300*time.Millisecond, 10*time.Millisecond, progress, cancel)
	defer stop()
	time.Sleep(100 * time.Millisecond)
	progress.Store(time.Now().UnixNano())
	time.Sleep(100 * time.Millisecond)
	if ctx.Err() != nil {
		t.Fatalf("watcher ignored fresh activity: %v", context.Cause(ctx))
	}
	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("idle worker was not cancelled")
	}
	if !errors.Is(context.Cause(ctx), agent.AbortWorkStalled) {
		t.Fatalf("cause=%v, want stalled", context.Cause(ctx))
	}
}

func TestWorkStallWatcherStopsWithoutCancelling(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	defer cancel(nil)
	progress := new(atomic.Int64)
	progress.Store(time.Now().UnixNano())
	stop := startWorkStallWatcher(ctx, time.Second, 10*time.Millisecond, progress, cancel)
	stop()
	if ctx.Err() != nil {
		t.Fatalf("stop cancelled a healthy run: %v", context.Cause(ctx))
	}
}
