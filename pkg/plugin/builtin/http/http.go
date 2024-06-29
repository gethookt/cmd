package http // import "hookt.dev/cmd/pkg/plugin/builtin/http"

import (
	"context"
	"log/slog"

	"hookt.dev/cmd/pkg/check"
	"hookt.dev/cmd/pkg/plugin/builtin/http/wire"
	"hookt.dev/cmd/pkg/proto"
)

type Plugin struct {
	wire.Config

	p *proto.P
}

func (p *Plugin) Name() string {
	return "http"
}

func New(opts ...func(*Plugin)) *Plugin {
	p := &Plugin{}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *Plugin) WithProto(q *proto.P) *Plugin {
	p.p = q
	return p
}

func (p *Plugin) Plugin(_ context.Context, q *proto.P) any {
	return p.WithProto(q)
}

func (p *Plugin) Init(_ context.Context, job *proto.Job) error {
	slog.Debug("http: init",
		"config", p.Config,
	)

	return nil
}

func (p *Plugin) Step(context.Context) any {
	return &Step{p: p}
}

type Step struct {
	wire.Step `json:",inline"`

	p *Plugin
}

func (s *Step) Run(_ context.Context, _ *check.S) error {
	return nil
}

func (s *Step) Stop() {}
