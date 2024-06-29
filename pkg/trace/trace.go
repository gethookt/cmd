package trace

import (
	"context"

	"hookt.dev/cmd/pkg/proto/wire"
)

var nop = &JobTrace{
	WireJob:    func(int, *wire.Job) {},
	WirePlugin: func(int, *wire.Plugin, any) {},
	WireStep:   func(int, *wire.Step, any) {},
	RunStep:    func() {},
	MatchStep:  func() {},
	TapMessage: func() {},
}

func WithJobTrace(ctx context.Context, trace *JobTrace) context.Context {
	return with(ctx, trace)
}

func ContextJobTrace(ctx context.Context) *JobTrace {
	if trace := from[JobTrace](ctx); trace != nil {
		return trace
	}
	return nop
}

type JobTrace struct {
	WireJob    func(index int, job *wire.Job)
	WirePlugin func(index int, plugin *wire.Plugin, impl any)
	WireStep   func(index int, step *wire.Step, impl any)
	RunStep    func()

	MatchStep  func()
	TapMessage func()
}

type EventInfo struct{}
