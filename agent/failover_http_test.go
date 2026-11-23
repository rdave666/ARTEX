package agent

import (
	"fmt"
	"github.com/Autumn-27/artex/llmpool"
	"github.com/Autumn-27/norma/llm"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// Real HTTP adapters, without contacting any saved/paid/public provider.
func TestHTTPProfileRotationAndCooldown(t *testing.T) {
	for _, stream := range []bool{false, true} {
		t.Run(fmt.Sprint(stream), func(t *testing.T) {
			var mu sync.Mutex
			calls := map[string]int{}
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				prefix := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")[0]
				mu.Lock()
				calls[prefix]++
				mu.Unlock()
				switch prefix {
				case "quota":
					w.WriteHeader(402)
					fmt.Fprint(w, `{"error":"insufficient credit"}`)
				case "limited":
					w.Header().Set("Retry-After", "600")
					w.WriteHeader(429)
					fmt.Fprint(w, `{"error":"rate limit"}`)
				case "down":
					w.WriteHeader(503)
					fmt.Fprint(w, `{"error":"temporarily unavailable"}`)
				case "ok":
					if stream {
						w.Header().Set("Content-Type", "application/x-ndjson")
						fmt.Fprintln(w, `{"message":{"role":"assistant","content":"OK"},"done":true,"done_reason":"stop"}`)
					} else {
						fmt.Fprint(w, `{"message":{"role":"assistant","content":"OK"},"done":true,"done_reason":"stop"}`)
					}
				default:
					t.Errorf("unexpected path %s", r.URL.Path)
					w.WriteHeader(404)
				}
			}))
			defer srv.Close()
			health := llmpool.NewRegistry(nil, nil)
			health.SetPolicy(1, 0)
			var members []*llmpool.Member
			for i, name := range []string{"quota", "limited", "down", "ok"} {
				cfg := Config{Format: FormatOllama, BaseURL: srv.URL + "/" + name, Model: "test", Retry: RetryConfig{ConnectAttempts: -1}}
				p, err := cfg.NewProvider()
				if err != nil {
					t.Fatal(err)
				}
				members = append(members, &llmpool.Member{ID: int64(i + 1), Name: name, Prov: p, Rank: 4 - i})
			}
			pool := llmpool.New(members, health)
			req := llm.CompletionRequest{Messages: []llm.Message{llm.UserText("Reply OK")}}
			for range 2 {
				if stream {
					acc := llm.NewAccumulator()
					for ev, err := range pool.Stream(t.Context(), req) {
						if err != nil {
							t.Fatal(err)
						}
						acc.Add(ev)
					}
					if acc.Message().Text() != "OK" {
						t.Fatal("missing reply")
					}
				} else {
					msg, _, _, err := pool.Complete(t.Context(), req)
					if err != nil || msg.Text() != "OK" {
						t.Fatalf("msg=%v err=%v", msg, err)
					}
				}
			}
			mu.Lock()
			defer mu.Unlock()
			for _, name := range []string{"quota", "limited", "down"} {
				if calls[name] != 1 {
					t.Errorf("%s called %d times; cooldown ignored", name, calls[name])
				}
			}
			if calls["ok"] != 2 {
				t.Errorf("healthy calls=%d", calls["ok"])
			}
			if time.Until(health.Get(2).OpenUntil) < 599*time.Second {
				t.Fatal("Retry-After600 not preserved")
			}
		})
	}
}
