package event // import "hookt.dev/cmd/pkg/plugin/builtin/event"

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"hookt.dev/cmd/pkg/check"
	"hookt.dev/cmd/pkg/errors"
	"hookt.dev/cmd/pkg/plugin/builtin/event/wire"
	"hookt.dev/cmd/pkg/proto"
	"hookt.dev/cmd/pkg/trace"
)

type step struct {
	c    chan proto.Message
	done chan struct{}
}

type Plugin struct {
	wire.Config

	p     *proto.P
	steps []step
	mux   chan proto.Message
	stop  chan int
}

func New(opts ...func(*Plugin)) *Plugin {
	p := &Plugin{
		mux:  make(chan proto.Message),
		stop: make(chan int),
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
	switch p.Config.Mode {
	case "", "async", "sync":
		// ok
	default:
		return errors.New("invalid mode %q", p.Config.Mode)
	}
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

	go p.schedule()

	slog.Debug("event: init",
		"config", p.Config,
	)

	return nil
}

func (p *Plugin) schedule() {
	for {
		select {
		case i := <-p.stop:
			s := &p.steps[i]
			close(s.done)
			s.done = nil
		case msg := <-p.mux:
			var steps []step
			for _, step := range p.steps {
				if step.done == nil {
					continue
				}
				steps = append(steps, step)
			}
			switch p.Config.Mode {
			case "sync":
				wg := Wait(msg)
				go func() {
					for _, step := range steps {
						select {
						case <-step.done:
							continue
						case step.c <- wg:
							if wg.Wait() {
								return
							}
						}
					}
				}()
			case "", "async":
				go func() {
					for _, step := range steps {
						select {
						case <-step.done:
							continue
						case step.c <- msg:
						}
					}
				}()
			}
		}
	}
}

func (p *Plugin) Step(ctx context.Context) any {
	s := step{
		c:    make(chan proto.Message),
		done: make(chan struct{}),
	}
	p.steps = append(p.steps, s)
	it, _ := time.ParseDuration(p.Config.InactiveTimeout)
	return &Step{
		i:  len(p.steps) - 1,
		p:  p,
		it: nonempty(it, 1*time.Minute),
	}
}

type Step struct {
	wire.Step

	i  int
	p  *Plugin
	it time.Duration
}

func group(ctx context.Context, name string) context.Context {
	return trace.With(ctx, "pattern-group", name)
}

func (s *Step) Run(ctx context.Context, c *check.S) error {
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
		case msg := <-s.step().c:
			if !inactive.Stop() {
				<-inactive.C
			}
			inactive.Reset(s.it)

			ctxt := ctx

			if i, ok := msg.(Indexer); ok {
				ctxt = trace.With(ctxt, "event-seq", strconv.Itoa(i.Index()))
			}

			obj := msg.Object()

			match, err := match.Match(group(ctxt, "match"), obj)
			if err != nil {
				return errors.New("failed to match on pattern: %w", err)
			}

			if !match {
				if wg, ok := msg.(WaitMessage); ok {
					wg.Done(false)
				}
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

			pass, err := pass.Match(group(ctxt, "pass"), obj)
			if err != nil {
				return errors.New("failed to match ok pattern: %w", err)
			}
			if wg, ok := msg.(WaitMessage); ok {
				wg.Done(pass)
			}
			if pass {
				c.OK()
				return nil
			}
		}
	}

	return nil
}

func (s *Step) Stop() {
	s.p.stop <- s.i
	s.drain()
}

func (s *Step) step() step {
	return s.p.steps[s.i]
}

func (s *Step) drain() {
	for {
		select {
		case _ = <-s.step().c:
			// drop the event
		case <-s.step().done:
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
