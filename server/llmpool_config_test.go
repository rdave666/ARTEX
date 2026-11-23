package server

import (
	"github.com/Autumn-27/artex/agent"
	"testing"
	"time"
)

func TestPoolProviderEstablishmentPolicy(t *testing.T) {
	skipWithoutDSN(t)
	m, err := NewManager(t.TempDir(), "")
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	s := &Server{m: m}
	original := agent.Config{Retry: agent.RetryConfig{ConnectAttempts: 5}, ResponseHeaderTimeout: 2 * time.Minute}
	if err := m.SetLLMPoolEnabled(false); err != nil {
		t.Fatal(err)
	}
	if got := s.poolProviderConfig(original); got != original {
		t.Fatal("exclusive provider policy changed")
	}
	if err := m.SetLLMPoolEnabled(true); err != nil {
		t.Fatal(err)
	}
	got := s.poolProviderConfig(original)
	if got.Retry.ConnectAttempts != -1 || got.ResponseHeaderTimeout != time.Minute {
		t.Fatalf("pool policy=%+v", got)
	}
	if original.Retry.ConnectAttempts != 5 || original.ResponseHeaderTimeout != 2*time.Minute {
		t.Fatal("source config mutated")
	}
}
