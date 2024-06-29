package proto // import "hookt.dev/cmd/pkg/proto"

import (
	"context"
	"encoding/json"
	"log/slog"

	"hookt.dev/cmd/pkg/errors"
	"hookt.dev/cmd/pkg/proto/wire"
	"hookt.dev/cmd/pkg/trace"

	"sigs.k8s.io/yaml"
)

type Workflow struct {
	Jobs []Job
}

type Job struct {
	Plugins []Plugin
	Steps   []Step
}

type Condition struct{}

type Step struct {
	Uses  string
	ID    string
	Desc  string
	With  any
	Ready context.Context // defer + timeout + wait_for
}

type Plugin struct {
	Uses string
	ID   string
	With any
}

type P struct {
	t *T
	m map[string]Interface
}

func New(opts ...func(*P)) *P {
	p := &P{
		t: &T{},
		m: make(map[string]Interface),
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *P) Parse(ctx context.Context, q []byte) (*Workflow, error) {
	raw, err := wire.XParse(q)
	if err != nil {
		return nil, errors.New("error parsing workflow: %w", err)
	}

	var (
		w  Workflow
		tr = trace.ContextJobTrace(ctx)
	)

	w.Jobs = make([]Job, len(raw.Jobs))

	for i, job := range raw.Jobs {
		j := &w.Jobs[i]

		slog.Debug("wiring jobs",
			"index", i,
			"job", job.ID,
		)

		tr.WireJob(i, &job)

		j.Plugins = make([]Plugin, len(job.Plugins))

		for k, plugin := range job.Plugins {
			iface, ok := p.m[plugin.Uses]
			if !ok {
				return nil, errors.New("error reading plugin %q config: not found", plugin.Uses)
			}

			slog.Debug("wiring plugins",
				"index", k,
				"plugin", plugin.Uses,
				"with", string(plugin.With),
			)

			q := &j.Plugins[k]
			q.ID = plugin.ID
			q.Uses = plugin.Uses
			q.With = iface.Plugin(ctx, p)

			if err := json.Unmarshal(plugin.With, q.With); err != nil {
				return nil, errors.New("error reading plugin %q config: %w", plugin.Uses, err)
			}

			tr.WirePlugin(k, &plugin, q.With)
		}

		j.Steps = make([]Step, len(job.Steps))

		for k, step := range job.Steps {
			iface, ok := p.m[step.Uses]
			if !ok {
				return nil, errors.New("error reading plugin %q step: not found", step.Uses)
			}

			slog.Debug("wiring steps",
				"index", k,
				"step", step.Uses,
				"with", string(step.With),
			)

			s := &j.Steps[k]

			s.Uses = step.Uses
			s.ID = step.ID
			s.Desc = step.Desc
			s.With = iface.Step(ctx)

			if err := yaml.Unmarshal(step.With, s.With); err != nil {
				return nil, errors.New("error reading plugin %q step: %w", step.Uses, err)
			}

			tr.WireStep(k, &step, s.With)
		}

		for k := range j.Plugins {
			p := &j.Plugins[k]

			slog.Debug("initializing plugins",
				"index", k,
				"plugin", p.Uses,
			)

			init, ok := p.With.(Initializer)
			if !ok {
				return nil, errors.New("error initializing plugin %q: does not implement proto.Initializer", p.Uses)
			}

			if err := init.Init(ctx, j); err != nil {
				return nil, errors.New("error initializing plugin %q: %w", p.Uses, err)
			}
		}
	}

	return &w, nil
}
