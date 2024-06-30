package proto // import "hookt.dev/cmd/pkg/proto"

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"
	"strings"

	"hookt.dev/cmd/pkg/errors"
	"hookt.dev/cmd/pkg/proto/wire"
	"hookt.dev/cmd/pkg/trace"

	"sigs.k8s.io/yaml"
)

type Workflow struct {
	Jobs []Job
}

type Job struct {
	ID      string
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
		tr = trace.ContextJob(ctx)
	)

	w.Jobs = make([]Job, len(raw.Jobs))

	uniq := make(map[string]struct{})

	for i, job := range raw.Jobs {
		j := &w.Jobs[i]

		if strings.HasPrefix(job.ID, "#") {
			return nil, errors.New("#job-%d: error reading job: id cannot start with #", i)
		}

		slog.Debug("wiring jobs",
			"index", i,
			"job", job.ID,
		)

		tr.WireJob(i, &job)

		j.ID = nonempty(job.ID, "#job-"+strconv.Itoa(i))
		j.Plugins = make([]Plugin, len(job.Plugins))

		if _, ok := uniq[j.ID]; ok {
			return nil, errors.New("#job-%d: error reading job: duplicate id %q", i, j.ID)
		}

		uniq[j.ID] = struct{}{}

		ctx := trace.With(ctx, "job", j.ID)

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

		uniq := make(map[string]struct{})

		for k, step := range job.Steps {
			iface, ok := p.m[step.Uses]
			if !ok {
				return nil, errors.New("#step-%d: error reading plugin %q step: not found", k, step.Uses)
			}

			if strings.HasPrefix(step.ID, "#") {
				return nil, errors.New("#step-%d: error reading plugin %q step: id cannot start with #", k, step.Uses)
			}

			s := &j.Steps[k]

			s.Uses = step.Uses
			s.ID = nonempty(step.ID, "#step-"+strconv.Itoa(k))
			s.Desc = step.Desc
			s.With = iface.Step(trace.With(ctx, "step", s.ID))

			if _, ok := uniq[s.ID]; ok {
				return nil, errors.New("error reading plugin %q step: duplicate id %q", step.Uses, s.ID)
			}

			uniq[s.ID] = struct{}{}

			if err := yaml.Unmarshal(step.With, s.With); err != nil {
				return nil, errors.New("%s: error reading plugin %q step: %w", s.ID, step.Uses, err)
			}

			slog.Debug("wiring steps",
				"id", s.ID,
				"step", step.Uses,
				"with", string(step.With),
			)

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

func nonempty[T comparable](t ...T) T {
	var zero T
	for _, v := range t {
		if v != zero {
			return v
		}
	}
	return zero
}
