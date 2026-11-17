package agent

import (
	"context"
	"iter"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Autumn-27/norma/llm"
	"github.com/Autumn-27/norma/transcript"
)

type modelChatProbe struct {
	requests                   []llm.CompletionRequest
	streamCalls, completeCalls int
}

func (p *modelChatProbe) Stream(_ context.Context, req llm.CompletionRequest) iter.Seq2[llm.StreamEvent, error] {
	p.requests = append(p.requests, req)
	p.streamCalls++
	return func(yield func(llm.StreamEvent, error) bool) {
		if yield(llm.StreamEvent{Type: llm.SETextDelta, Text: "hello"}, nil) {
			yield(llm.StreamEvent{Type: llm.SEMessageDelta, StopReason: "end_turn"}, nil)
		}
	}
}
func (p *modelChatProbe) Complete(_ context.Context, req llm.CompletionRequest) (llm.Message, string, llm.Usage, error) {
	p.requests = append(p.requests, req)
	p.completeCalls++
	return llm.Message{Role: llm.RoleAssistant, Content: []llm.ContentBlock{llm.TextBlock("hello")}}, "end_turn", llm.Usage{}, nil
}
func TestModelChatLightweightAndHistory(t *testing.T) {
	for _, nonStreaming := range []bool{false, true} {
		t.Run(map[bool]string{false: "stream", true: "complete"}[nonStreaming], func(t *testing.T) {
			p := &modelChatProbe{}
			dir := t.TempDir()
			c := NewChatAgent(p, "test-model", dir, transcript.NewStore(filepath.Join(dir, "transcripts")), 100000)
			c.SetNonStreaming(func() bool { return nonStreaming })
			c.SetWebSearch(WebSearchOpts{Enabled: true})
			for _, msg := range []string{"first question", "second question"} {
				answer, err := c.Chat(t.Context(), "model_chat", "conv-model_chat", msg, 1, 0, true, nil)
				if err != nil || answer != "hello" {
					t.Fatalf("answer=%q err=%v", answer, err)
				}
			}
			if len(p.requests) != 2 {
				t.Fatalf("requests=%d", len(p.requests))
			}
			for _, req := range p.requests {
				if len(req.Tools) != 0 {
					t.Errorf("model chat advertised %d tools", len(req.Tools))
				}
				system := strings.Join(req.System, "\n")
				if len(system) > 1024 {
					t.Errorf("model chat system bloated: %d bytes", len(system))
				}
				if strings.Contains(system, "working directory") || strings.Contains(system, "noise") {
					t.Errorf("agent instructions leaked: %q", system)
				}
			}
			// Norma adds a date reminder; the real conversation must still resume
			// intact after it rather than resetting to the newest user message.
			history := p.requests[1].Messages
			if len(history) < 3 {
				t.Fatalf("history messages=%d", len(history))
			}
			history = history[len(history)-3:]
			if history[0].Text() != "first question" || history[1].Role != llm.RoleAssistant || history[1].Text() != "hello" || history[2].Text() != "second question" {
				t.Fatalf("conversation history not preserved: %+v", history)
			}

			if nonStreaming && p.completeCalls != 2 || !nonStreaming && p.streamCalls != 2 {
				t.Fatalf("stream=%d complete=%d", p.streamCalls, p.completeCalls)
			}
		})
	}
}
func TestModelChatLeavesNormalAgentToolsEnabled(t *testing.T) {
	p := &modelChatProbe{}
	c := NewChatAgent(p, "test-model", t.TempDir(), nil, 100000)
	if _, err := c.Chat(t.Context(), "auto", "conv-normal-chat", "hello", 1, 0, false, nil); err != nil {
		t.Fatal(err)
	}
	if len(p.requests) != 1 || len(p.requests[0].Tools) == 0 {
		t.Fatal("normal chat lost its tools")
	}
}
