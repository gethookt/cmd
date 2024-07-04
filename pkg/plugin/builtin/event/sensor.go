package event

import (
	"context"
	"text/template"

	"hookt.dev/cmd/pkg/async"
	"hookt.dev/cmd/pkg/errors"
	"hookt.dev/cmd/pkg/plugin/builtin/event/wire"
	"hookt.dev/cmd/pkg/proto"
)

type Locals struct {
	async.Map
}

func (l *Locals) setlocal(name string, value any) any {
	l.Map.Store(name, value)
	return value
}

func (l *Locals) getlocal(name string) any {
	value, _ := l.Map.Load(name)
	return value
}

func (l *Locals) opts() []proto.TOption {
	return []proto.TOption{
		func(t *template.Template) *template.Template {
			return t.Funcs(template.FuncMap{
				"setlocal": l.setlocal,
				"local":    l.getlocal,
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
