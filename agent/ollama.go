package agent

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Autumn-27/norma/llm"
	"github.com/google/uuid"
)

// ollamaProvider uses the native API because OpenAI compatibility cannot set
// num_ctx. Changing ARTEX's local compaction window alone never changes the
// server's inference context allocation.
type ollamaProvider struct {
	cfg       Config
	client    *http.Client
	nextReply atomic.Uint64
	identity  string
}

func newOllamaProvider(c Config, client *http.Client) llm.Provider {
	c.Format = FormatOllama
	return &ollamaProvider{cfg: c, client: client, identity: uuid.NewString()}
}

type ollamaFunction struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}
type ollamaToolCall struct {
	Function ollamaFunction `json:"function"`
}
type ollamaMessage struct {
	Role      string           `json:"role"`
	Content   string           `json:"content"`
	Thinking  string           `json:"thinking,omitempty"`
	ToolCalls []ollamaToolCall `json:"tool_calls,omitempty"`
	ToolName  string           `json:"tool_name,omitempty"`
}
type ollamaChunk struct {
	Message         ollamaMessage `json:"message"`
	Done            bool          `json:"done"`
	DoneReason      string        `json:"done_reason"`
	PromptEvalCount int           `json:"prompt_eval_count"`
	EvalCount       int           `json:"eval_count"`
	Error           string        `json:"error"`
}

func ollamaNativeURL(base string) string {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	// ConfigFrom historically appends /v1 even to a pasted native endpoint.
	// Strip both layers without discarding an actual reverse-proxy prefix.
	for {
		previous := base
		for _, suffix := range []string{"/v1/chat/completions", "/api/chat", "/api/generate", "/api/tags", "/v1"} {
			if strings.HasSuffix(base, suffix) {
				base = strings.TrimSuffix(base, suffix)
				break
			}
		}
		if previous == base {
			break
		}
	}
	if base == "" {
		base = "http://127.0.0.1:11434"
	}
	return base + "/api/chat"
}

func (p *ollamaProvider) requestBody(req llm.CompletionRequest, stream bool) ([]byte, error) {
	messages := make([]ollamaMessage, 0, len(req.Messages)+1)
	if system := strings.Join(req.System, "\n\n"); system != "" {
		messages = append(messages, ollamaMessage{Role: "system", Content: system})
	}
	names := map[string]string{}
	for _, m := range req.Messages {
		native := ollamaMessage{Role: string(m.Role)}
		for _, b := range m.Content {
			switch b.Type {
			case llm.BlockText:
				native.Content += b.Text
			case llm.BlockThinking:
				native.Thinking += b.Thinking
			case llm.BlockToolUse:
				arguments := b.Input
				if len(arguments) == 0 {
					arguments = json.RawMessage(`{}`)
				}
				var object map[string]json.RawMessage
				if err := json.Unmarshal(arguments, &object); err != nil || object == nil {
					return nil, fmt.Errorf("ollama: tool %q arguments must be a JSON object", b.Name)
				}
				native.ToolCalls = append(native.ToolCalls, ollamaToolCall{Function: ollamaFunction{Name: b.Name, Arguments: arguments}})
				names[b.ID] = b.Name
			case llm.BlockToolResult:
				if native.Content != "" {
					messages = append(messages, native)
					native = ollamaMessage{Role: string(m.Role)}
				}
				name, ok := names[b.ToolUseID]
				if !ok {
					return nil, fmt.Errorf("ollama: tool result has no matching tool call %q", b.ToolUseID)
				}
				result := ""
				for _, part := range b.Content {
					if part.Type == llm.BlockText {
						result += part.Text
					}
				}
				messages = append(messages, ollamaMessage{Role: "tool", Content: result, ToolName: name})
			}
		}
		if native.Content != "" || native.Thinking != "" || len(native.ToolCalls) > 0 {
			messages = append(messages, native)
		}
	}
	// Allocate exactly the same effective context used by local compaction.
	options := map[string]any{"num_ctx": p.cfg.CompactionWindow()}
	if req.MaxTokens > 0 {
		options["num_predict"] = req.MaxTokens
	}
	if req.Temperature != nil {
		options["temperature"] = *req.Temperature
	}
	if len(req.Stop) > 0 {
		options["stop"] = req.Stop
	}
	body := map[string]any{"model": p.cfg.Model, "messages": messages, "options": options, "stream": stream, "truncate": false}
	if len(req.Tools) > 0 {
		tools := make([]map[string]any, 0, len(req.Tools))
		for _, tool := range req.Tools {
			tools = append(tools, map[string]any{"type": "function", "function": map[string]any{"name": tool.Name, "description": tool.Description, "parameters": tool.InputSchema}})
		}
		body["tools"] = tools
	}
	thinking := p.cfg.ThinkingType
	if req.Thinking != "" {
		thinking = req.Thinking
	}
	switch thinking {
	case "disabled":
		body["think"] = false
	case "enabled":
		if p.cfg.ReasoningEffort != "" {
			body["think"] = p.cfg.ReasoningEffort
		} else {
			body["think"] = true
		}
	default:
		if p.cfg.ReasoningEffort != "" {
			body["think"] = p.cfg.ReasoningEffort
		}
	}
	return json.Marshal(body)
}

func (p *ollamaProvider) send(ctx context.Context, req llm.CompletionRequest, stream bool) (*http.Response, error) {
	body, err := p.requestBody(req, stream)
	if err != nil {
		return nil, err
	}
	retries := p.cfg.Retry.ConnectAttempts
	if retries == 0 {
		retries = 3
	}
	if retries < 0 {
		retries = 0
	}
	for attempt := 0; ; attempt++ {
		request, err := http.NewRequestWithContext(ctx, http.MethodPost, ollamaNativeURL(p.cfg.BaseURL), bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("ollama: build request: %w", err)
		}
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Accept", "application/json, application/x-ndjson")
		if p.cfg.APIKey != "" {
			request.Header.Set("Authorization", "Bearer "+p.cfg.APIKey)
		}
		response, err := p.client.Do(request)
		retry := err != nil && ctx.Err() == nil
		if response != nil {
			switch response.StatusCode {
			case 408, 429, 500, 502, 503, 504:
				retry = true
			}
		}
		if retry && attempt < retries {
			if response != nil {
				_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
				response.Body.Close()
			}
			delay := p.cfg.Retry.ConnectInterval
			if delay <= 0 {
				delay = 500 * time.Millisecond * time.Duration(1<<min(attempt, 4))
			}
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return nil, ctx.Err()
			case <-timer.C:
			}
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("ollama: request failed: %w", err)
		}
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			raw, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
			response.Body.Close()
			detail := string(raw)
			if p.cfg.APIKey != "" {
				detail = strings.ReplaceAll(detail, p.cfg.APIKey, "<REDACTED>")
			}
			return nil, fmt.Errorf("ollama: status %d: %s", response.StatusCode, detail)
		}
		return response, nil
	}
}

func (p *ollamaProvider) Complete(ctx context.Context, req llm.CompletionRequest) (llm.Message, string, llm.Usage, error) {
	response, err := p.send(ctx, req, false)
	if err != nil {
		return llm.Message{}, "", llm.Usage{}, err
	}
	defer response.Body.Close()
	const limit = 16 << 20
	raw, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err != nil {
		return llm.Message{}, "", llm.Usage{}, err
	}
	if len(raw) > limit {
		return llm.Message{}, "", llm.Usage{}, fmt.Errorf("ollama: response exceeds 16MiB")
	}
	var chunk ollamaChunk
	if err = json.Unmarshal(raw, &chunk); err != nil {
		return llm.Message{}, "", llm.Usage{}, fmt.Errorf("ollama: invalid completion JSON: %w", err)
	}
	acc := llm.NewAccumulator()
	seq := p.nextReply.Add(1)
	toolIndex := 0
	if err = p.emitChunk(chunk, seq, &toolIndex, func(ev llm.StreamEvent) bool { acc.Add(ev); return true }); err != nil {
		return llm.Message{}, "", llm.Usage{}, err
	}
	if !chunk.Done {
		return llm.Message{}, "", llm.Usage{}, fmt.Errorf("ollama: incomplete completion (missing done:true)")
	}
	return acc.Message(), acc.StopReason, acc.Usage, nil
}

func (p *ollamaProvider) Stream(ctx context.Context, req llm.CompletionRequest) iter.Seq2[llm.StreamEvent, error] {
	return func(yield func(llm.StreamEvent, error) bool) {
		response, err := p.send(ctx, req, true)
		if err != nil {
			yield(llm.StreamEvent{}, err)
			return
		}
		defer response.Body.Close()
		if !yield(llm.StreamEvent{Type: llm.SEMessageStart}, nil) {
			return
		}
		scanner := bufio.NewScanner(response.Body)
		scanner.Buffer(make([]byte, 64<<10), 16<<20)
		seq := p.nextReply.Add(1)
		toolIndex := 0
		stopped := false
		for scanner.Scan() {
			if len(bytes.TrimSpace(scanner.Bytes())) == 0 {
				continue
			}
			var chunk ollamaChunk
			if err := json.Unmarshal(scanner.Bytes(), &chunk); err != nil {
				yield(llm.StreamEvent{}, fmt.Errorf("ollama: invalid stream JSON: %w", err))
				return
			}
			err = p.emitChunk(chunk, seq, &toolIndex, func(ev llm.StreamEvent) bool {
				if stopped {
					return false
				}
				stopped = !yield(ev, nil)
				return !stopped
			})
			if err != nil {
				yield(llm.StreamEvent{}, err)
				return
			}
			if stopped || chunk.Done {
				return
			}
		}
		if err := scanner.Err(); err != nil {
			yield(llm.StreamEvent{}, fmt.Errorf("ollama: stream read: %w", err))
			return
		}
		yield(llm.StreamEvent{}, fmt.Errorf("ollama: stream ended before done:true"))
	}
}

func (p *ollamaProvider) emitChunk(chunk ollamaChunk, seq uint64, toolIndex *int, emit func(llm.StreamEvent) bool) error {
	if chunk.Error != "" {
		detail := chunk.Error
		if p.cfg.APIKey != "" {
			detail = strings.ReplaceAll(detail, p.cfg.APIKey, "<REDACTED>")
		}
		return fmt.Errorf("ollama: %s", detail)
	}
	if chunk.Message.Thinking != "" && !emit(llm.StreamEvent{Type: llm.SEThinkingDelta, Text: chunk.Message.Thinking}) {
		return nil
	}
	if chunk.Message.Content != "" && !emit(llm.StreamEvent{Type: llm.SETextDelta, Text: chunk.Message.Content}) {
		return nil
	}
	for _, tool := range chunk.Message.ToolCalls {
		var object map[string]json.RawMessage
		if tool.Function.Name == "" || json.Unmarshal(tool.Function.Arguments, &object) != nil || object == nil {
			return fmt.Errorf("ollama: invalid tool call name or arguments")
		}
		id := fmt.Sprintf("ollama_%s_%d_%d", p.identity, seq, *toolIndex)
		if !emit(llm.StreamEvent{Type: llm.SEToolUseStart, ToolID: id, ToolName: tool.Function.Name, ToolIndex: *toolIndex}) {
			return nil
		}
		*toolIndex++
		if !emit(llm.StreamEvent{Type: llm.SEToolInputJSON, Text: string(tool.Function.Arguments)}) {
			return nil
		}
	}
	if chunk.Done {
		stop := "end_turn"
		if *toolIndex > 0 {
			stop = "tool_use"
		} else if chunk.DoneReason == "length" {
			stop = "max_tokens"
		}
		if !emit(llm.StreamEvent{Type: llm.SEMessageDelta, StopReason: stop, Usage: llm.Usage{InputTokens: chunk.PromptEvalCount, OutputTokens: chunk.EvalCount}}) {
			return nil
		}
		emit(llm.StreamEvent{Type: llm.SEMessageStop})
	}
	return nil
}
