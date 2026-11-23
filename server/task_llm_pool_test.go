package server

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Autumn-27/artex/db"
	"github.com/Autumn-27/artex/llmpool"
	"github.com/Autumn-27/norma/llm"
)

// No engine or real LLM is started: only routing and persisted task selection
// are exercised against the disposable test database.
func newTaskPoolFixture(t *testing.T) (*Server, *taskLLMRuntime, *Task, int64, *scriptedLLMProvider, *scriptedLLMProvider) {
	t.Helper()
	skipWithoutDSN(t)
	m, err := NewManager(t.TempDir(), "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { m.Close() })
	oldEnabled, oldFallback := m.LLMPoolEnabled(), m.LLMPoolBindFallback()
	t.Cleanup(func() { _ = m.SetLLMPoolEnabled(oldEnabled); _ = m.SetLLMPoolBindFallback(oldFallback) })
	ids := make([]int64, 2)
	for i := range ids {
		ids[i], err = m.pg.SaveProfile(&db.LLMProfile{Name: fmt.Sprintf("task-pool-%d-%d", time.Now().UnixNano(), i), Format: "ollama", Model: "fake"})
		if err != nil {
			t.Fatal(err)
		}
		id := ids[i]
		t.Cleanup(func() { _ = m.pg.DeleteProfile(id) })
	}
	task, err := m.CreateTaskWithOptions("pool routing fixture", "do not execute", db.TaskCreateOptions{LLMProfileIDs: []int64{ids[0]}})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = m.DeleteTask(task.ID, DeleteTaskOptions{}) })
	head := &scriptedLLMProvider{err: errors.New("openai: status 429: rate limit")}
	backup := &scriptedLLMProvider{events: []llm.StreamEvent{{Type: llm.SETextDelta, Text: "fallback"}}}
	cache := map[int64]*provEntry{}
	profiles, err := m.pg.PoolProfiles()
	if err != nil {
		t.Fatal(err)
	}
	// NewManager may seed example profiles; never build a real provider for
	// them. The only successful endpoint in this test is the fake backup.
	for _, profile := range profiles {
		cache[profile.ID] = &provEntry{prov: &scriptedLLMProvider{err: errors.New("fixture endpoint unreachable")}}
	}
	cache[ids[0]], cache[ids[1]] = &provEntry{prov: head}, &provEntry{prov: backup}
	s := &Server{m: m, llmOn: true, llmProv: backup, llmHealth: llmpool.NewRegistry(nil, nil), provByProfile: cache}
	return s, &taskLLMRuntime{s: s, taskID: task.ID}, task, ids[0], head, backup
}

func TestTaskSelectedProfilePoolFallback(t *testing.T) {
	for _, tc := range []struct {
		name              string
		enabled, fallback bool
		failure           string
	}{
		{"enabled_rate_limit", true, true, "openai: status 429: rate limit"},
		{"enabled_quota", true, true, "openai: status 402: insufficient_quota"},
		{"enabled_unreachable", true, true, "dial tcp: connection refused"},
		{"master_off", false, true, "openai: status 429: rate limit"},
		{"bound_fallback_off", true, false, "openai: status 429: rate limit"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, runtime, _, headID, head, backup := newTaskPoolFixture(t)
			head.err = errors.New(tc.failure)
			if err := s.m.SetLLMPoolEnabled(tc.enabled); err != nil {
				t.Fatal(err)
			}
			if err := s.m.SetLLMPoolBindFallback(tc.fallback); err != nil {
				t.Fatal(err)
			}
			before, err := s.m.pg.GetTask(mustTaskID(t, runtime.taskID))
			if err != nil {
				t.Fatal(err)
			}
			selection, err := runtime.currentFor(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			events, streamErr := collectTaskLLMStream(selection.provider.Stream(context.Background(), llm.CompletionRequest{}))
			if tc.enabled && tc.fallback {
				if streamErr != nil || len(events) != 1 || events[0].Text != "fallback" || backup.calls != 1 {
					t.Fatalf("events=%v err=%v backup calls=%d", events, streamErr, backup.calls)
				}
			} else if streamErr == nil || backup.calls != 0 {
				t.Fatalf("exclusive binding fell back: err=%v backup calls=%d", streamErr, backup.calls)
			}
			if head.calls != 1 || selection.profileID != headID {
				t.Fatalf("head calls=%d selected=%d", head.calls, selection.profileID)
			}
			after, err := s.m.pg.GetTask(before.ID)
			if err != nil {
				t.Fatal(err)
			}
			if after.ActiveLLMProfileID == nil || *after.ActiveLLMProfileID != headID || after.LLMChainRevision != before.LLMChainRevision || len(after.LLMProfileIDs) != 1 {
				t.Fatal("pool fallback mutated persisted task chain")
			}
		})
	}
}

func TestTaskExhaustedChainCanUseGlobalPoolFallback(t *testing.T) {
	for _, tc := range []struct {
		name              string
		enabled, fallback bool
	}{
		{"enabled", true, true}, {"master_off", false, true}, {"bound_fallback_off", true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, runtime, task, headID, head, backup := newTaskPoolFixture(t)
			if err := s.m.SetLLMPoolEnabled(tc.enabled); err != nil {
				t.Fatal(err)
			}
			if err := s.m.SetLLMPoolBindFallback(tc.fallback); err != nil {
				t.Fatal(err)
			}
			id := mustTaskID(t, task.ID)
			if _, err := s.m.pg.MarkTaskLLMProfileQuotaExhausted(id, headID, "insufficient_quota"); err != nil {
				t.Fatal(err)
			}
			before, err := s.m.pg.GetTask(id)
			if err != nil {
				t.Fatal(err)
			}
			s.syncTaskLLMState(before)
			wantReady := tc.enabled && tc.fallback
			if ready := s.taskRuntimeAvailable(task, "unbound-test-agent"); ready != wantReady {
				t.Errorf("runtime available=%v want %v", ready, wantReady)
			}
			selection, err := runtime.currentFor(context.Background())
			if wantReady {
				if err != nil {
					t.Fatal(err)
				}
				if selection.profileID != 0 || selection.provider != backup {
					t.Fatal("exhausted chain did not use global provider")
				}
				if _, err := collectTaskLLMStream(selection.provider.Stream(context.Background(), llm.CompletionRequest{})); err != nil {
					t.Fatal(err)
				}
			} else if err == nil {
				t.Fatal("exhausted exclusive chain unexpectedly resolved")
			}
			after, err := s.m.pg.GetTask(id)
			if err != nil {
				t.Fatal(err)
			}
			if after.ActiveLLMProfileID != nil || after.LLMChainRevision != before.LLMChainRevision || after.LLMFailoverState != before.LLMFailoverState || len(after.LLMProfileIDs) != 1 || head.calls != 0 {
				t.Fatal("global fallback altered or retried the exhausted task chain")
			}
		})
	}
}
