package trace

import (
	"context"
	"fmt"
	"log/slog"
	"text/template"

	"hookt.dev/cmd/pkg/proto/wire"

	"github.com/itchyny/gojq"
	"github.com/lmittmann/tint"
)

var (
	nopJob = JobTrace{
		WireJob:    func(int, *wire.Job) {},
		WirePlugin: func(int, *wire.Plugin, any) {},
		WireStep:   func(int, *wire.Step, any) {},
		RunStep:    func() {},
		MatchStep:  func() {},
		TapMessage: func() {},
	}
	nopPattern = PatternTrace{
		ParseKey:       func(string, *gojq.Query, error) {},
		UnmarshalValue: func(string, []byte, any, error) {},
		TemplateValue:  func(string, string, *template.Template, error) {},
		ExecuteMatch:   func(string, []byte, error) {},
		UnmarshalMatch: func(string, []byte, any, error) {},
		EqualMatch:     func(string, any, any, bool) {},
	}
)

func LogJob() JobTrace {
	return JobTrace{
		WireJob:    func(int, *wire.Job) {},
		WirePlugin: func(int, *wire.Plugin, any) {},
		WireStep:   func(int, *wire.Step, any) {},
		RunStep:    func() {},
		MatchStep:  func() {},
		TapMessage: func() {},
	}
}

func LogPattern() PatternTrace {
	return PatternTrace{
		ParseKey: func(key string, q *gojq.Query, err error) {
			if err != nil {
				slog.Error("trace: ParseKey",
					"key", key,
					tint.Err(err),
				)
			} else {
				slog.Info("trace: ParseKey",
					"key", key,
				)
			}
		},
		UnmarshalValue: func(key string, p []byte, v any, err error) {
			if err != nil {
				slog.Error("trace: UnmarshalValue",
					"key", key,
					"raw", string(p),
					"value", v,
					tint.Err(err),
				)
			} else {
				slog.Info("trace: UnmarshalValue",
					"key", key,
					"raw", string(p),
					"value", v,
				)
			}
		},
		TemplateValue: func(key string, value string, t *template.Template, err error) {
			if err != nil {
				slog.Error("trace: TemplateValue",
					"key", key,
					"value", value,
					tint.Err(err),
				)
			} else {
				slog.Info("trace: TemplateValue",
					"key", key,
					"value", value,
				)
			}
		},
		ExecuteMatch: func(key string, p []byte, err error) {
			if err != nil {
				slog.Error("trace: ExecuteMatch",
					"key", key,
					"raw", string(p),
					tint.Err(err),
				)
			} else {
				slog.Info("trace: ExecuteMatch",
					"key", key,
					"raw", string(p),
				)
			}
		},
		UnmarshalMatch: func(key string, p []byte, v any, err error) {
			if err != nil {
				slog.Error("trace: UnmarshalMatch",
					"key", key,
					"raw", string(p),
					"value", v,
					tint.Err(err),
				)
			} else {
				slog.Info("trace: UnmarshalMatch",
					"key", key,
					"raw", string(p),
					"value", v,
				)
			}
		},
		EqualMatch: func(key string, want any, got any, ok bool) {
			if !ok {
				slog.Error("trace: EqualMatch",
					"key", key,
					"want", fmt.Sprintf("%+[1]v (%[1]T)", want),
					"got", fmt.Sprintf("%+[1]v (%[1]T)", got),
				)
			} else {
				slog.Info("trace: EqualMatch",
					"key", key,
					"want", fmt.Sprintf("%+[1]v (%[1]T)", want),
					"got", fmt.Sprintf("%+[1]v (%[1]T)", got),
				)
			}
		},
	}
}

func WithJob(ctx context.Context, trace JobTrace) context.Context {
	return with(ctx, &trace)
}

func ContextJob(ctx context.Context) JobTrace {
	if trace := from[JobTrace](ctx); trace != nil {
		return *trace
	}
	return nopJob
}

func WithPattern(ctx context.Context, trace PatternTrace) context.Context {
	return with(ctx, &trace)
}

func ContextPattern(ctx context.Context) PatternTrace {
	if trace := from[PatternTrace](ctx); trace != nil {
		return *trace
	}
	return nopPattern
}

type JobTrace struct {
	WireJob    func(index int, job *wire.Job)
	WirePlugin func(index int, plugin *wire.Plugin, impl any)
	WireStep   func(index int, step *wire.Step, impl any)
	RunStep    func()

	MatchStep  func()
	TapMessage func()
}

type PatternTrace struct {
	ParseKey       func(string, *gojq.Query, error)
	UnmarshalValue func(string, []byte, any, error)
	TemplateValue  func(string, string, *template.Template, error)
	ExecuteMatch   func(string, []byte, error)
	UnmarshalMatch func(string, []byte, any, error)
	EqualMatch     func(string, any, any, bool)
}

type EventInfo struct{}
