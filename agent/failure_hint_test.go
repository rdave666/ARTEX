package agent

import (
	"context"
	"errors"
	"github.com/Autumn-27/norma/llm"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestProviderPreservesRetryAfter(t *testing.T) {
	for _, format := range []llm.Format{FormatOllama, llm.FormatOpenAI} {
		t.Run(string(format), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Retry-After", "120")
				w.WriteHeader(429)
				w.Write([]byte(`{"error":"rate limit"}`))
			}))
			defer srv.Close()
			cfg := Config{Format: format, BaseURL: srv.URL, Model: "test", APIKey: "test", Retry: RetryConfig{ConnectAttempts: -1}}
			p, err := cfg.NewProvider()
			if err != nil {
				t.Fatal(err)
			}
			_, _, _, err = p.Complete(t.Context(), llm.CompletionRequest{Messages: []llm.Message{llm.UserText("hi")}})
			var hint interface {
				RetryAfter() time.Duration
				StatusCode() int
			}
			if !errors.As(err, &hint) || hint.StatusCode() != 429 || hint.RetryAfter() != 120*time.Second {
				t.Fatalf("missing HTTP hint: %v", err)
			}
		})
	}
}
func TestProviderHeaderTimeoutIsLocal(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(time.Second):
		}
	}))
	defer srv.Close()
	cfg := Config{Format: FormatOllama, BaseURL: srv.URL, Model: "test", Retry: RetryConfig{ConnectAttempts: -1}, ResponseHeaderTimeout: 20 * time.Millisecond}
	p, err := cfg.NewProvider()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()
	_, _, _, err = p.Complete(ctx, llm.CompletionRequest{Messages: []llm.Message{llm.UserText("hi")}})
	if err == nil || ctx.Err() != nil {
		t.Fatalf("want provider-local timeout, got %v parent=%v", err, ctx.Err())
	}
}
func TestParseRetryAfter(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	for _, tc := range []struct {
		s    string
		want time.Duration
	}{{"60", time.Minute}, {now.Add(2 * time.Minute).Format(http.TimeFormat), 2 * time.Minute}, {"-1", 0}, {"bad", 0}, {"", 0}} {
		if got := parseRetryAfter(tc.s, now); got != tc.want {
			t.Errorf("%s: %v want %v", tc.s, got, tc.want)
		}
	}
}

func TestProviderPreservesCreditFailureFromHTTP400(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		w.Write([]byte(`{"error":"insufficient credit balance"}`))
	}))
	defer srv.Close()
	cfg := Config{Format: FormatOllama, BaseURL: srv.URL, Model: "test", Retry: RetryConfig{ConnectAttempts: -1}}
	p, err := cfg.NewProvider()
	if err != nil {
		t.Fatal(err)
	}
	_, _, _, err = p.Complete(t.Context(), llm.CompletionRequest{})
	var hint interface{ StatusCode() int }
	if !errors.As(err, &hint) || hint.StatusCode() != 402 {
		t.Fatalf("credit exhaustion must be distinguished from malformed request: %v", err)
	}
}
