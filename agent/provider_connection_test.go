package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// The connection test must test the model, not silently install the SDK's
// default filesystem/shell tools (nil Tools means DefaultTools in agentcore).
func TestConnectionIsLightweightAndNeverRunsTools(t *testing.T) {
	for _, stream := range []bool{false, true} {
		t.Run(fmt.Sprint(stream), func(t *testing.T) {
			var body map[string]json.RawMessage
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
					return
				}
				if stream {
					w.Header().Set("Content-Type", "application/x-ndjson")
					fmt.Fprintln(w, `{"message":{"role":"assistant","content":"OK"},"done":false}`)
					fmt.Fprintln(w, `{"message":{"role":"assistant","content":""},"done":true,"done_reason":"stop"}`)
				} else {
					w.Header().Set("Content-Type", "application/json")
					fmt.Fprint(w, `{"message":{"role":"assistant","content":"OK"},"done":true,"done_reason":"stop"}`)
				}
			}))
			defer upstream.Close()
			cfg := ConfigFrom("ollama", "small-model", upstream.URL, "", "")
			cfg.Stream = stream
			cfg.MaxTokens = 128
			_, reply, err := TestConnection(context.Background(), cfg)
			if err != nil || reply != "OK" {
				t.Fatalf("reply=%q err=%v", reply, err)
			}
			if tools := body["tools"]; len(tools) > 0 && string(tools) != "[]" {
				t.Errorf("connection test unexpectedly advertised tools (%d bytes)", len(tools))
			}
			var options struct {
				NumPredict int `json:"num_predict"`
				NumCtx     int `json:"num_ctx"`
			}
			if err := json.Unmarshal(body["options"], &options); err != nil || options.NumPredict != 128 || options.NumCtx != 32000 {
				t.Errorf("test ignored native context/output cap: %s (%v)", body["options"], err)
			}
			if len(body["messages"]) > 1024 {
				t.Errorf("unexpected agent prompt injected: %d bytes", len(body["messages"]))
			}
			if string(body["stream"]) != fmt.Sprint(stream) {
				t.Errorf("stream setting ignored: %s", body["stream"])
			}
		})
	}
}
