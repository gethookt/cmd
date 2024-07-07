package event

import (
	"context"
	"sync"
	"text/template"

	"hookt.dev/cmd/pkg/errors"
	"hookt.dev/cmd/pkg/plugin/builtin/event/wire"
	"hookt.dev/cmd/pkg/proto"
)

type Tags struct {
	mu sync.Mutex
	m  map[string]any
}

func MakeTags() *Tags {
	return &Tags{
		m: make(map[string]any),
	}
}

func (t *Tags) tag(name string, value ...any) (any, error) {
	switch len(value) {
	case 0:
		t.mu.Lock()
		value, ok := t.m[name]
		t.mu.Unlock()
		if !ok {
			return nil, errors.New("tag not found: %q", name)
		}
		return value, nil
	case 1:
		t.mu.Lock()
		t.m[name] = value
		t.mu.Unlock()
		return value, nil
	default:
		return nil, errors.New("too many arguments for tag: %q", name)
	}
}

func (t *Tags) opts() []proto.TOption {
	return []proto.TOption{
		func(tmpl *template.Template) *template.Template {
			return tmpl.Funcs(template.FuncMap{
				"tag": t.tag,
			})
		},
	}
}

type Sensor struct {
	Match proto.Patterns
	Pass  proto.Patterns
	Fail  proto.Patterns
}

func (p *Plugin) MakeSensor(ctx context.Context, step *wire.Step, opts ...proto.TOption) (*Sensor, error) {
	var (
		s   Sensor
		err error
	)

	s.Match, err = p.p.Patterns(group(ctx, "match"), step.Match, opts...)
	if err != nil {
		return nil, errors.New("failed to parse match pattern: %w", err)
	}

	s.Fail, err = p.p.Patterns(group(ctx, "fail"), step.Fail, opts...)
	if err != nil {
		return nil, errors.New("failed to parse fail pattern: %w", err)
	}

	s.Pass, err = p.p.Patterns(group(ctx, "pass"), step.Pass, opts...)
	if err != nil {
		return nil, errors.New("failed to parse pass pattern: %w", err)
	}

	return &s, nil
}

func (s *Sensor) Do(ctx context.Context, obj any) (bool, error) {
	match, err := s.Match.Match(group(ctx, "match"), obj)
	if err != nil {
		return false, errors.New("failed to match on pattern: %w", err)
	}
	if !match {
		return false, nil
	}

	fail, err := s.Fail.Match(group(ctx, "fail"), obj)
	if err != nil {
		return false, errors.New("failed to match fail pattern: %w", err)
	}
	if fail && len(s.Fail) != 0 {
		return false, errors.New("failure pattern matched")
	}

	pass, err := s.Pass.Match(group(ctx, "pass"), obj)
	if err != nil {
		return false, errors.New("failed to match ok pattern: %w", err)
	}

	return pass, nil
}
