package llmpool

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/Autumn-27/norma/llm"
)

// Member is one LLM profile in the chain, already built into a provider (wrapped
// with the recorder by the caller, so a failed attempt is still recorded under
// its own profile name).
type Member struct {
	ID       int64  // llm_profiles.id
	Name     string // profile name, for logs / UI
	Model    string
	Format   string // "anthropic" | "openai"
	Priority int    // the profile's configured priority (display only)
	Active   bool   // is_default (display only)
	// Rank is the ordering key the caller assigned: higher goes first, and members
	// sharing a Rank take turns leading (load-spreading across duplicate keys).
	// The caller encodes "active profile heads the chain" as a Rank above every
	// user-settable priority, so this type needs no policy of its own.
	Rank int
	// WindowTokens is the profile's context window in tokens. A member whose
	// window can't hold the request is skipped rather than made to fail on it.
	WindowTokens int
	Prov         llm.Provider
}

// RankActive is the Rank the caller gives the chain head (the active profile, or
// an explicitly bound one) so it always outranks any configured priority.
const RankActive = int(^uint(0)>>1) - 1

// ErrExhausted is returned when every member of the chain failed.
var ErrExhausted = errors.New("LLM failover: all profiles unavailable")

// Pool is an llm.Provider that fails over across an ordered chain of members.
// It is safe for concurrent use: members are immutable after construction and
// all mutable state lives in the shared Registry.
type Pool struct {
	members []*Member // in chain order (active first, then priority DESC)
	health  *Registry
	rr      atomic.Uint64 // rotates the starting point within an equal-priority group
}

// New builds a Pool over members (already in chain order). Returns nil when the
// chain is empty. A single-member chain is still a valid Pool — it just behaves
// exactly like the bare provider.
func New(members []*Member, health *Registry) *Pool {
	if len(members) == 0 {
		return nil
	}
	if health == nil {
		health = NewRegistry(nil, nil)
	}
	return &Pool{members: members, health: health}
}

// Members returns the chain in order (read-only).
func (p *Pool) Members() []*Member { return p.members }

// Head returns the first member of the chain.
func (p *Pool) Head() *Member { return p.members[0] }

// Stream implements llm.Provider with failover.
//
// Only semantic output commits an attempt. Provider metadata is buffered until
// content arrives, so an error after message_start can still rotate safely.
// Once text, thinking or a tool call reaches the caller, failover is forbidden.
func (p *Pool) Stream(ctx context.Context, req llm.CompletionRequest) iter.Seq2[llm.StreamEvent, error] {
	return func(yield func(llm.StreamEvent, error) bool) {
		if err := ctx.Err(); err != nil {
			yield(llm.StreamEvent{}, err)
			return
		}
		order := p.order(req)
		var lastErr error
		for i, m := range order {
			if err := ctx.Err(); err != nil {
				yield(llm.StreamEvent{}, err)
				return
			}
			if p.health.IsOpen(m.ID) {
				continue
			}
			committed := false
			var pending []llm.StreamEvent
			flush := func() bool {
				for _, ev := range pending {
					if !yield(ev, nil) {
						return false
					}
				}
				pending = nil
				return true
			}
			var failed error
			for ev, err := range m.Prov.Stream(ctx, req) {
				if err != nil {
					if !committed && shouldFailover(ctx, err) {
						failed = err
						break
					}
					if !flush() {
						return
					}
					yield(ev, err)
					return
				}
				if !committed && !semanticEvent(ev) {
					// Bound buffering if a broken endpoint emits only metadata.
					if len(pending) >= 64 {
						failed = errors.New("LLM stream sent too much metadata before any content")
						break
					}
					pending = append(pending, ev)
					continue
				}
				if !committed {
					committed = true
					if !flush() {
						return
					}
				}
				if !yield(ev, nil) {
					return
				}
			}
			if failed == nil {
				p.health.RecordCall(ctx, m.ID, true, false, nil) // completed: Pass
				return
			}
			lastErr = failed
			if tripped, msg := p.health.RecordCall(ctx, m.ID, true, false, failed); tripped {
				log.Printf("[llmpool] profile %q(%s) breaker tripped: %s", m.Name, m.Model, msg)
			}
			if i+1 < len(order) {
				n := order[i+1]
				log.Printf("[llmpool] LLM failover: %q(%s) → %q(%s), reason: %s", m.Name, m.Model, n.Name, n.Model, trimErr(failed))
			}
		}
		if lastErr == nil {
			lastErr = ErrExhausted
		}
		yield(llm.StreamEvent{}, fmt.Errorf("%w: %v", ErrExhausted, lastErr))
	}
}

// Metadata and empty deltas do not expose assistant content. Unknown event
// types conservatively commit the attempt rather than risk replaying effects.
func semanticEvent(ev llm.StreamEvent) bool {
	switch ev.Type {
	case llm.SEMessageStart, llm.SEMessageDelta, llm.SEMessageStop:
		return false
	case llm.SETextDelta, llm.SEThinkingDelta, llm.SEThinkingSignature, llm.SERedactedThinking, llm.SEToolInputJSON, llm.SEToolUseSignature:
		return ev.Text != ""
	default:
		return true
	}
}

// Complete implements llm.Provider with failover for non-streaming calls. A
// non-streaming request is atomic — it never delivers partial output — so every
// failover-eligible failure lands in the safe window and the next member can be
// tried without risk of duplicated output. Mirrors Stream's health-tripping and
// chain-exhaustion behavior.
func (p *Pool) Complete(ctx context.Context, req llm.CompletionRequest) (llm.Message, string, llm.Usage, error) {
	if err := ctx.Err(); err != nil {
		return llm.Message{}, "", llm.Usage{}, err
	}
	order := p.order(req)
	var lastErr error
	for i, m := range order {
		if err := ctx.Err(); err != nil {
			return llm.Message{}, "", llm.Usage{}, err
		}
		if p.health.IsOpen(m.ID) {
			continue
		}
		msg, sr, usage, err := m.Prov.Complete(ctx, req)
		if err == nil {
			p.health.RecordCall(ctx, m.ID, false, false, nil) // Pass
			return msg, sr, usage, nil
		}
		if !shouldFailover(ctx, err) {
			// Non-failover error (e.g. ctx cancel, deterministic 4xx): surface as-is
			// without tripping health. RecordCall clears the state on an atomic
			// call, matching the historical Pass here.
			p.health.RecordCall(ctx, m.ID, false, false, err)
			return llm.Message{}, "", llm.Usage{}, err
		}
		lastErr = err
		if tripped, msg := p.health.RecordCall(ctx, m.ID, false, false, err); tripped {
			log.Printf("[llmpool] profile %q(%s) breaker tripped: %s", m.Name, m.Model, msg)
		}
		if i+1 < len(order) {
			n := order[i+1]
			log.Printf("[llmpool] LLM failover: %q(%s) → %q(%s), reason: %s",
				m.Name, m.Model, n.Name, n.Model, trimErr(err))
		}
	}
	if lastErr == nil {
		lastErr = ErrExhausted
	}
	log.Printf("[llmpool] failover chain exhausted (all %d profiles failed), last error: %s", len(order), trimErr(lastErr))
	return llm.Message{}, "", llm.Usage{}, fmt.Errorf("%w: %v", ErrExhausted, lastErr)
}

// order picks the members to try, in order: skip those in a cooling-off window
// and those whose context window can't hold this request, then rotate within each
// equal-priority group so same-priority profiles share the load. If all profiles
// are cooling down or too small, return no candidates instead of violating
// provider cooldowns or sending a request that is known not to fit.
func (p *Pool) order(req llm.CompletionRequest) []*Member {
	est := estimateTokens(req)
	var open []*Member
	for _, m := range p.members {
		if p.health.IsOpen(m.ID) {
			continue
		}
		if m.WindowTokens > 0 && est > m.WindowTokens {
			continue // would 400 on length — not a useful failover target
		}
		open = append(open, m)
	}
	return rotateGroups(open, p.rr.Add(1)-1)
}

// rotateGroups rotates each run of equal-ranked members by n, so profiles sharing
// a priority take turns going first (free load-spreading across duplicate keys).
// The chain head has its own Rank and always stays at the front.
func rotateGroups(in []*Member, n uint64) []*Member {
	out := make([]*Member, 0, len(in))
	for i := 0; i < len(in); {
		j := i + 1
		for j < len(in) && in[j].Rank == in[i].Rank {
			j++
		}
		g := in[i:j]
		if len(g) > 1 {
			off := int(n % uint64(len(g)))
			for k := range g {
				out = append(out, g[(k+off)%len(g)])
			}
		} else {
			out = append(out, g...)
		}
		i = j
	}
	return out
}

// statusRe pulls the HTTP status out of the SDK's error text, which is formatted
// as "<prefix>: status <code>: <body>" (norma/llm/retry.go). The SDK exposes no
// typed error, so the string is what we have.
var statusRe = regexp.MustCompile(`status (\d{3})`)

// statusOf returns the HTTP status carried by err, or 0 if it isn't one.
func statusOf(err error) int {
	var status interface{ StatusCode() int }
	if errors.As(err, &status) {
		return status.StatusCode()
	}
	m := statusRe.FindStringSubmatch(err.Error())
	if m == nil {
		return 0
	}
	code, _ := strconv.Atoi(m[1])
	return code
}

// shouldFailover reports whether err justifies trying the next profile.
//
// Never fails over on context cancellation — that's the user stopping a task or a
// task-level timeout, and burning a backup key on it would both waste credit and
// pollute the run's termination diagnosis. Never on 400 either: a malformed or
// over-long request fails identically everywhere.
func shouldFailover(ctx context.Context, err error) bool {
	if err == nil {
		return false
	}
	if ctx.Err() != nil || errors.Is(err, context.Canceled) {
		return false
	}
	switch code := statusOf(err); {
	case code == 0:
		return true // no status → transport-level failure (reset / DNS / timeout)
	case code == 400:
		return false // bad or over-long request: identical everywhere
	case code == 401, code == 402, code == 403, code == 404, code == 408, code == 429:
		return true
	case code >= 500:
		return true
	}
	return false
}

// retryAfter extracts a provider's Retry-After hint without depending on its
// concrete error type. Wrapped transport/SDK errors retain the hint.
func retryAfter(err error) time.Duration {
	var hint interface{ RetryAfter() time.Duration }
	if errors.As(err, &hint) {
		return max(hint.RetryAfter(), 0)
	}
	return 0
}

// isHardFailure reports whether the failure is deterministic (the profile will
// keep failing until a human fixes it) rather than transient. Hard failures open
// the breaker on the first occurrence.
func isHardFailure(err error) bool {
	switch statusOf(err) {
	case 401, 402, 403, 404:
		return true
	}
	return false
}

// trimErr shortens an error for logs/UI — provider bodies can be long.
func trimErr(err error) string {
	s := strings.TrimSpace(strings.ReplaceAll(err.Error(), "\n", " "))
	if len(s) > 300 {
		cut := 300
		for cut > 0 && !utf8.RuneStart(s[cut]) {
			cut--
		}
		s = s[:cut] + "…"
	}
	return s
}

// estimateTokens roughly sizes a request so members whose context window clearly
// can't hold it are skipped. Deliberately crude (~3.5 chars/token) and biased to
// over-estimate slightly; it only needs to tell "fits" from "nowhere near".
func estimateTokens(req llm.CompletionRequest) int {
	n := 0
	for _, s := range req.System {
		n += len(s)
	}
	for _, m := range req.Messages {
		n += blocksLen(m.Content)
	}
	for _, t := range req.Tools {
		// InputSchema is a decoded map, so its serialized size isn't available
		// cheaply — charge a flat ~120 chars per top-level property instead.
		n += len(t.Name) + len(t.Description) + len(t.InputSchema)*120
	}
	return n * 2 / 7 // ≈ len/3.5
}

func blocksLen(bs []llm.ContentBlock) int {
	n := 0
	for _, b := range bs {
		n += len(b.Text) + len(b.Thinking) + len(b.Input) + len(b.Name)
		if len(b.Content) > 0 {
			n += blocksLen(b.Content)
		}
	}
	return n
}
