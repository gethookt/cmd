package check // import "hookt.dev/cmd/pkg/check"

import (
	"context"
	"strconv"
	"sync"

	"github.com/itchyny/gojq"
	"hookt.dev/cmd/pkg/trace"
)

type S struct {
	mu sync.Mutex

	Events Events

	Steps struct {
		OK   int
		Fail int
	}
}

func (s *S) OK() {
	s.mu.Lock()
	s.Steps.OK++
	s.mu.Unlock()
}

func (s *S) Fail() {
	s.mu.Lock()
	s.Steps.Fail++
	s.mu.Unlock()
}

func (s *S) Trace() trace.PatternTrace {
	return trace.PatternTrace{
		ParseKey: func(ctx context.Context, q *gojq.Query, err error) {
			if err != nil {
				return
			}

			n, _ := strconv.Atoi(trace.Get(ctx, "step-index"))
			desc := trace.Get(ctx, "step-desc")
			group := trace.Get(ctx, "pattern-group")
			pattern := trace.Get(ctx, "pattern")

			s.mu.Lock()

			if n >= len(s.Events) {
				for i := len(s.Events); i <= n; i++ {
					s.Events = append(s.Events, Event{})
				}
			}

			s.Events[n].Desc = desc
			s.Events[n].MarkPattern(group, pattern, false)

			s.mu.Unlock()
		},
		EqualMatch: func(ctx context.Context, a, b any, eq bool) {
			if !eq {
				return
			}

			n, _ := strconv.Atoi(trace.Get(ctx, "step-index"))
			group := trace.Get(ctx, "pattern-group")
			pattern := trace.Get(ctx, "pattern")

			s.mu.Lock()

			s.Events[n].MarkPattern(group, pattern, eq)

			s.mu.Unlock()
		},
	}
}

func (e *Event) MarkPattern(group, pattern string, ok bool) {
	switch group {
	case "match":
		if e.Match == nil {
			e.Match = make(map[string]bool)
		}
		e.Match[pattern] = ok
	case "pass":
		if e.Pass == nil {
			e.Pass = make(map[string]bool)
		}
		e.Pass[pattern] = ok
	case "fail":
		if e.Fail == nil {
			e.Fail = make(map[string]bool)
		}
		e.Fail[pattern] = ok
	default:
		panic("unknown group: " + group)
	}
}

type Event struct {
	Desc  string          `json:"desc,omitempty"`
	Match map[string]bool `json:"match"`
	Pass  map[string]bool `json:"pass,omitempty"`
	Fail  map[string]bool `json:"fail,omitempty"`
}

type Events []Event
