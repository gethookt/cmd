package event // import "hookt.dev/cmd/pkg/plugin/builtin/event"

import (
	"context"
	"log/slog"
	"maps"
	"strconv"
	"time"

	"hookt.dev/cmd/pkg/check"
	"hookt.dev/cmd/pkg/errors"
	"hookt.dev/cmd/pkg/plugin/builtin/event/wire"
	"hookt.dev/cmd/pkg/proto"
	"hookt.dev/cmd/pkg/trace"
)

type Plugin struct {
	wire.Config

	p    *proto.P
	c    map[chan proto.Message]chan struct{}
	mux  chan proto.Message
	stop chan (chan proto.Message)
}

func New(opts ...func(*Plugin)) *Plugin {
	p := &Plugin{
		c:    make(map[chan proto.Message]chan struct{}),
		mux:  make(chan proto.Message),
		stop: make(chan chan proto.Message),
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *Plugin) WithProto(q *proto.P) *Plugin {
	p.p = q
	return p
}

func (p *Plugin) Name() string {
	return "event"
}

func (p *Plugin) Plugin(_ context.Context, q *proto.P) any {
	return p.WithProto(q)
}

func (p *Plugin) Init(ctx context.Context, job *proto.Job) error {
wire:
	for _, source := range p.Config.Sources {
		for _, plugin := range job.Plugins {
			if source != plugin.ID {
				continue
			}

			sub, ok := plugin.With.(proto.Subscriber)
			if !ok {
				return errors.New("plugin %q does not implement proto.Subscriber", plugin.Uses)
			}

			go func() {
				for msg := range sub.Subscribe(ctx) {
					p.mux <- msg
				}
			}()

			continue wire
		}

		return errors.New("source %q not found in job plugins", source)
	}

	go p.process()

	slog.Debug("event: init",
		"config", p.Config,
	)

	return nil
}

func (p *Plugin) process() {
	type indexer interface {
		Index() int
	}

	for {
		select {
		case c := <-p.stop:
			done := p.c[c]
			close(done)
			delete(p.c, c)
		case msg := <-p.mux:
			ch := maps.Clone(p.c)
			go func() {
				for c, done := range ch {
					select {
					case <-done:
					case c <- msg:
					}
				}
			}()
		}
	}
}

func (p *Plugin) Step(ctx context.Context) any {
	c := make(chan proto.Message)
	done := make(chan struct{})
	p.c[c] = done
	it, _ := time.ParseDuration(p.Config.InactiveTimeout)
	return &Step{
		p:    p,
		c:    c,
		done: done,
		it:   nonempty(it, 1*time.Minute),
	}
}

type Step struct {
	wire.Step

	p    *Plugin
	c    chan proto.Message
	done chan struct{}
	it   time.Duration
}

func group(ctx context.Context, name string) context.Context {
	return trace.With(ctx, "pattern-group", name)
}

func (s *Step) Run(ctx context.Context, c *check.S) error {
	type indexer interface {
		Index() int
	}

	slog.Debug("event: run",
		"match", s.Match,
		"pass", s.Pass,
		"fail", s.Fail,
	)

	tr := trace.ContextPattern(ctx)

	match, err := s.p.p.Patterns(group(ctx, "match"), s.Match)
	if err != nil {
		return errors.New("failed to parse match pattern: %w", err)
	}

	pass, err := s.p.p.Patterns(group(ctx, "pass"), s.Pass)
	if err != nil {
		return errors.New("failed to parse pass pattern: %w", err)
	}

	fail, err := s.p.p.Patterns(group(ctx, "fail"), s.Fail)
	if err != nil {
		return errors.New("failed to parse fail pattern: %w", err)
	}

	inactive := time.NewTimer(s.it)
	defer inactive.Stop()

	for {
		select {
		case <-inactive.C:
			c.Fail()
			tr.MatchTimeout(ctx)
			return errors.New("step has timed out after %v", s.it)
		case msg := <-s.c:
			if !inactive.Stop() {
				<-inactive.C
			}
			inactive.Reset(s.it)

			ctxt := ctx

			if i, ok := msg.(indexer); ok {
				ctxt = trace.With(ctxt, "event-seq", strconv.Itoa(i.Index()))
			}

			obj := msg.Object()

			match, err := match.Match(group(ctxt, "match"), obj)
			if err != nil {
				return errors.New("failed to match on pattern: %w", err)
			}

			if !match {
				continue
			}

			fail, err := fail.Match(group(ctxt, "fail"), obj)
			if err != nil {
				return errors.New("failed to match fail pattern: %w", err)
			}
			if fail {
				c.Fail()
				return errors.New("failure pattern matched")
			}

			ok, err := pass.Match(group(ctxt, "pass"), obj)
			if err != nil {
				return errors.New("failed to match ok pattern: %w", err)
			}
			if ok {
				c.OK()
				return nil
			}
		}
	}

	return nil
}

func (s *Step) Stop() {
	s.p.stop <- s.c
	s.drain()
}

func (s *Step) drain() {
	for {
		select {
		case <-s.c:
		case <-s.done:
			return
		}
	}
}

func nonempty[T comparable](t ...T) T {
	var zero T
	for _, v := range t {
		if v != zero {
			return v
		}
	}
	return zero
}
