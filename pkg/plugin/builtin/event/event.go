package event // import "hookt.dev/cmd/pkg/plugin/builtin/event"

import (
	"context"
	"fmt"
	"log/slog"
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
	c    map[chan proto.Message]string
	mux  chan proto.Message
	stop chan (chan proto.Message)
}

func New(opts ...func(*Plugin)) *Plugin {
	p := &Plugin{
		c:    make(map[chan proto.Message]string),
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
			delete(p.c, c)
			close(c)
		case msg := <-p.mux:
			go func() {
				for c, id := range p.c {
					fmt.Printf("\nDEBUG: MUXING EVENT TO %q: seq=%d\n\n", id, msg.(indexer).Index())
					c <- msg
					fmt.Printf("\nDEBUG: MUXED EVENT TO %q: seq=%d\n\n", id, msg.(indexer).Index())
				}
			}()
		}
	}
}

func (p *Plugin) Step(ctx context.Context) any {
	c := make(chan proto.Message)
	p.c[c] = trace.Get(ctx, "step")
	it, _ := time.ParseDuration(p.Config.InactiveTimeout)
	return &Step{
		p:  p,
		c:  c,
		it: nonempty(it, 1*time.Minute),
	}
}

type Step struct {
	wire.Step

	p  *Plugin
	c  chan proto.Message
	it time.Duration
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

	step := trace.Get(ctx, "step")

	defer fmt.Printf("\nDEBUG: RUN DONE IN %q\n\n", step)

	for {
		fmt.Printf("\nDEBUG: WAITING FOR EVENT IN %q\n\n", step)

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

				fmt.Printf("\nDEBUG: EVENT RECEIVED IN %q: seq=%d\n\n", step, i.Index())
			} else {
				fmt.Printf("\nDEBUG: EVENT RECEIVED IN %q\n\n", step)
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
	fmt.Printf("\nDEBUG: STOPPING STEP\n\n")
	s.p.stop <- s.c
	fmt.Printf("\nDEBUG: STOPPED STEP\n\n")
	s.drain()
	fmt.Printf("\nDEBUG: DRAINED STEP\n\n")
}

func (s *Step) drain() {
	for range s.c {
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
