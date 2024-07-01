package check // import "hookt.dev/cmd/pkg/check"

import (
	"context"
	"fmt"
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

func (s *S) Results() []Result {
	var res []Result
	for i, e := range s.Events {
		f := makeFailures(e.Match, false)
		if len(f) > 0 {
			res = append(res, Result{
				Type:     "match",
				Index:    i,
				Step:     e.Desc,
				Failures: f,
			})
			continue
		}

		f = makeFailures(e.Fail, true)
		if len(f) > 0 {
			res = append(res, Result{
				Type:     "fail",
				Index:    i,
				Step:     e.Desc,
				Failures: f,
			})
			continue
		}

		f = makeFailures(e.Pass, false)
		if len(f) > 0 {
			res = append(res, Result{
				Type:     "pass",
				Index:    i,
				Step:     e.Desc,
				Failures: f,
			})
			continue
		}
	}
	return res
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
			defer s.mu.Unlock()

			if n >= len(s.Events) {
				for i := len(s.Events); i <= n; i++ {
					s.Events = append(s.Events, &Event{})
				}
			}

			s.Events[n].Desc = desc
			s.Events[n].MarkPattern(group, pattern, Value{
				OK: false,
			})
		},
		EqualMatch: func(ctx context.Context, want, got any, ok bool) {
			n, _ := strconv.Atoi(trace.Get(ctx, "step-index"))
			group := trace.Get(ctx, "pattern-group")
			pattern := trace.Get(ctx, "pattern")

			s.mu.Lock()
			defer s.mu.Unlock()

			s.Events[n].MarkPattern(group, pattern, Value{
				Want: want,
				Got:  got,
				OK:   ok,
			})
		},
	}
}

func (e *Event) MarkPattern(group, pattern string, v Value) {
	switch group {
	case "match":
		if e.Match == nil {
			e.Match = make(map[string]Value)
		}
		e.Match[pattern] = v
	case "pass":
		if e.Pass == nil {
			e.Pass = make(map[string]Value)
		}
		e.Pass[pattern] = v
	case "fail":
		if e.Fail == nil {
			e.Fail = make(map[string]Value)
		}
		e.Fail[pattern] = v
	default:
		panic(fmt.Errorf("unknown group: %q (pattern=%q, ok=%v)", group, pattern, v))
	}
}

type Value struct {
	Got  any  `json:"got,omitempty"`
	Want any  `json:"want,omitempty"`
	OK   bool `json:"ok"`
}

type Event struct {
	Desc  string           `json:"desc,omitempty"`
	Match map[string]Value `json:"match"`
	Pass  map[string]Value `json:"pass,omitempty"`
	Fail  map[string]Value `json:"fail,omitempty"`
}

type Events []*Event

type Result struct {
	Type     string    `json:"type"`
	Step     string    `json:"step"`
	Index    int       `json:"index"`
	Failures []Failure `json:"failures"`
}

type Failure struct {
	Key      string `json:"key"`
	Got      any    `json:"got"`
	Expected any    `json:"expected"`
}

func makeFailures(m map[string]Value, ok bool) []Failure {
	var failures []Failure
	for key, v := range m {
		if v.OK == ok {
			failures = append(failures, Failure{
				Key:      key,
				Got:      v.Got,
				Expected: v.Want,
			})
		}
	}
	return failures
}
