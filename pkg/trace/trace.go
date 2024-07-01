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
		Step:       func(string) StepTrace { return nopStep },
	}
	nopStep = StepTrace{
		PatternGroup: func(string, string) PatternGroupTrace { return nopPatternGroup },
	}
	nopPatternGroup = PatternGroupTrace{
		Pattern: func(string) PatternTrace { return nopPattern },
	}
	nopPattern = PatternTrace{
		ParseKey:       func(context.Context, *gojq.Query, error) {},
		UnmarshalValue: func(context.Context, []byte, any, error) {},
		TemplateValue:  func(context.Context, string, *template.Template, error) {},
		ExecuteMatch:   func(context.Context, []byte, error) {},
		UnmarshalMatch: func(context.Context, []byte, any, error) {},
		EqualMatch:     func(context.Context, any, any, bool) {},
		MatchTimeout:   func(context.Context) {},
	}
	nopSchedule = ScheduleTrace{
		BeforePublish: func(context.Context, *wire.Message) {},
		Publish:       func(context.Context, *wire.Message) {},
		BeforeStop:    func(context.Context, int) {},
		Stop:          func(context.Context, int) {},
		BeforeDemux:   func(context.Context, Message) {},
		Demux:         func(context.Context, Message) {},
		BeforeMux:     func(context.Context, Message, int) {},
		Mux:           func(context.Context, Message, int) {},
		Wait:          func(context.Context, Message, int, bool) {},
		Done:          func(context.Context, Message, int) {},
		Drain:         func(context.Context, int) {},
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
		ParseKey: func(ctx context.Context, q *gojq.Query, err error) {
			tags := append(attrs(ctx))
			if err != nil {
				tags = append(tags, tint.Err(err))
				slog.Error("trace: ParseKey", tags...)
			} else {
				slog.Info("trace: ParseKey", tags...)
			}
		},
		UnmarshalValue: func(ctx context.Context, p []byte, v any, err error) {
			tags := append(attrs(ctx),
				"raw", string(p),
				"value", v,
			)
			if err != nil {
				tags = append(tags, tint.Err(err))
				slog.Error("trace: UnmarshalValue", tags...)
			} else {
				slog.Info("trace: UnmarshalValue", tags...)
			}
		},
		TemplateValue: func(ctx context.Context, value string, t *template.Template, err error) {
			tags := append(attrs(ctx),
				"value", value,
			)
			if err != nil {
				tags = append(tags, tint.Err(err))
				slog.Error("trace: TemplateValue", tags...)
			} else {
				slog.Info("trace: TemplateValue", tags...)
			}
		},
		ExecuteMatch: func(ctx context.Context, p []byte, err error) {
			tags := append(attrs(ctx),
				"raw", string(p),
			)
			if err != nil {
				tags = append(tags, tint.Err(err))
				slog.Error("trace: ExecuteMatch", tags...)
			} else {
				slog.Info("trace: ExecuteMatch", tags...)
			}
		},
		UnmarshalMatch: func(ctx context.Context, p []byte, v any, err error) {
			tags := append(attrs(ctx),
				"raw", string(p),
				"value", v,
			)
			if err != nil {
				tags = append(tags, tint.Err(err))
				slog.Error("trace: UnmarshalMatch", tags...)
			} else {
				slog.Info("trace: UnmarshalMatch", tags...)
			}
		},
		EqualMatch: func(ctx context.Context, want any, got any, ok bool) {
			tags := append(attrs(ctx),
				slog.Group("want",
					"value", want,
					"type", fmt.Sprintf("%T", want),
				),
				slog.Group("got",
					"value", got,
					"type", fmt.Sprintf("%T", got),
				),
			)
			if !ok {
				slog.Error("trace: EqualMatch", tags...)
			} else {
				slog.Info("trace: EqualMatch", tags...)
			}
		},
		MatchTimeout: func(ctx context.Context) {
			tags := append(attrs(ctx))
			slog.Error("trace: MatchTimeout", tags...)
		},
	}
}

func LogSchedule() ScheduleTrace {
	return ScheduleTrace{
		BeforePublish: func(ctx context.Context, msg *wire.Message) {
			tags := append(attrs(ctx),
				"message", len(msg.Bytes()),
			)
			slog.Info("trace: BeforePublish", tags...)
		},
		Publish: func(ctx context.Context, msg *wire.Message) {
			tags := append(attrs(ctx),
				"message", len(msg.Bytes()),
			)
			slog.Info("trace: Publish", tags...)
		},
		BeforeStop: func(ctx context.Context, i int) {
			tags := append(attrs(ctx),
				"step", i,
			)
			slog.Info("trace: BeforeStop", tags...)
		},
		Stop: func(ctx context.Context, i int) {
			tags := append(attrs(ctx),
				"step", i,
			)
			slog.Info("trace: Stop", tags...)
		},
		BeforeDemux: func(ctx context.Context, msg Message) {
			tags := append(attrs(ctx),
				"message", len(msg.Bytes()),
			)
			slog.Info("trace: BeforeDemux", tags...)
		},
		Demux: func(ctx context.Context, msg Message) {
			tags := append(attrs(ctx),
				"message", len(msg.Bytes()),
			)
			slog.Info("trace: Demux", tags...)
		},
		BeforeMux: func(ctx context.Context, msg Message, i int) {
			tags := append(attrs(ctx),
				"message", len(msg.Bytes()),
				"step", i,
			)
			slog.Info("trace: BeforeMux", tags...)
		},
		Mux: func(ctx context.Context, msg Message, i int) {
			tags := append(attrs(ctx),
				"message", len(msg.Bytes()),
				"step", i,
			)
			slog.Info("trace: Mux", tags...)
		},
		Wait: func(ctx context.Context, msg Message, i int, ok bool) {
			tags := append(attrs(ctx),
				"message", len(msg.Bytes()),
				"step", i,
			)
			if !ok {
				slog.Error("trace: Wait", tags...)
			} else {
				slog.Info("trace: Wait", tags...)
			}
		},
		Done: func(ctx context.Context, msg Message, i int) {
			tags := append(attrs(ctx),
				"message", len(msg.Bytes()),
				"step", i,
			)
			slog.Info("trace: Done", tags...)
		},
		Drain: func(ctx context.Context, i int) {
			tags := append(attrs(ctx),
				"step", i,
			)
			slog.Info("trace: Drain", tags...)
		},
	}
}

func attrs(ctx context.Context) []any {
	var attrs []any

	if seq := Get(ctx, "event-seq"); seq != "" {
		attrs = append(attrs, "event-seq", seq)
	}
	if job := Get(ctx, "job"); job != "" {
		attrs = append(attrs, "job", job)
	}
	if step := Get(ctx, "step"); step != "" {
		attrs = append(attrs, "step", step)
	}
	if group := Get(ctx, "pattern-group"); group != "" {
		attrs = append(attrs, "pattern-group", group)
	}
	if pattern := Get(ctx, "pattern"); pattern != "" {
		attrs = append(attrs, "pattern", pattern)
	}

	return attrs
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

func WithStep(ctx context.Context, trace StepTrace) context.Context {
	return with(ctx, &trace)
}

func ContextStep(ctx context.Context) StepTrace {
	if trace := from[StepTrace](ctx); trace != nil {
		return *trace
	}
	return nopStep
}

func WithPatternGroup(ctx context.Context, trace PatternGroupTrace) context.Context {
	return with(ctx, &trace)
}

func ContextPatternGroup(ctx context.Context) PatternGroupTrace {
	if trace := from[PatternGroupTrace](ctx); trace != nil {
		return *trace
	}
	return nopPatternGroup
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

func WithSchedule(ctx context.Context, trace ScheduleTrace) context.Context {
	return with(ctx, &trace)
}

func ContextSchedule(ctx context.Context) ScheduleTrace {
	if trace := from[ScheduleTrace](ctx); trace != nil {
		return *trace
	}
	return nopSchedule
}

type traceKey struct{ string }

func With(ctx context.Context, key, value string) context.Context {
	return context.WithValue(ctx, traceKey{key}, value)
}

func Get(ctx context.Context, key string) string {
	value, _ := ctx.Value(traceKey{key}).(string)
	return value
}

type Trace struct{}

type JobTrace struct {
	WireJob    func(index int, job *wire.Job)
	WirePlugin func(index int, plugin *wire.Plugin, impl any)
	WireStep   func(index int, step *wire.Step, impl any)
	RunStep    func()

	MatchStep  func()
	TapMessage func()

	Step func(id string) StepTrace
}

type StepTrace struct {
	PatternGroup func(step, name string) PatternGroupTrace
}

type PatternGroupTrace struct {
	Pattern func(key string) PatternTrace
}

type Message interface {
	Bytes() []byte
	Object() any
}

type ScheduleTrace struct {
	BeforePublish func(context.Context, *wire.Message)
	Publish       func(context.Context, *wire.Message)
	BeforeStop    func(context.Context, int)
	Stop          func(context.Context, int)
	BeforeDemux   func(context.Context, Message)
	Demux         func(context.Context, Message)
	BeforeMux     func(context.Context, Message, int)
	Mux           func(context.Context, Message, int)
	Wait          func(context.Context, Message, int, bool)
	Done          func(context.Context, Message, int)
	Drain         func(context.Context, int)
}

type PatternTrace struct {
	ParseKey       func(context.Context, *gojq.Query, error)
	UnmarshalValue func(context.Context, []byte, any, error)
	TemplateValue  func(context.Context, string, *template.Template, error)
	ExecuteMatch   func(context.Context, []byte, error)
	UnmarshalMatch func(context.Context, []byte, any, error)
	EqualMatch     func(context.Context, any, any, bool)
	MatchTimeout   func(context.Context)
}

func (pt PatternTrace) Join(extra PatternTrace) PatternTrace {
	if extra.ParseKey != nil {
		fn := pt.ParseKey
		pt.ParseKey = func(ctx context.Context, q *gojq.Query, err error) {
			fn(ctx, q, err)
			extra.ParseKey(ctx, q, err)
		}
	}
	if extra.UnmarshalValue != nil {
		fn := pt.UnmarshalValue
		pt.UnmarshalValue = func(ctx context.Context, p []byte, v any, err error) {
			fn(ctx, p, v, err)
			extra.UnmarshalValue(ctx, p, v, err)
		}
	}
	if extra.TemplateValue != nil {
		fn := pt.TemplateValue
		pt.TemplateValue = func(ctx context.Context, value string, t *template.Template, err error) {
			fn(ctx, value, t, err)
			extra.TemplateValue(ctx, value, t, err)
		}
	}
	if extra.ExecuteMatch != nil {
		fn := pt.ExecuteMatch
		pt.ExecuteMatch = func(ctx context.Context, p []byte, err error) {
			fn(ctx, p, err)
			extra.ExecuteMatch(ctx, p, err)
		}
	}
	if extra.UnmarshalMatch != nil {
		fn := pt.UnmarshalMatch
		pt.UnmarshalMatch = func(ctx context.Context, p []byte, v any, err error) {
			fn(ctx, p, v, err)
			extra.UnmarshalMatch(ctx, p, v, err)
		}
	}
	if extra.EqualMatch != nil {
		fn := pt.EqualMatch
		pt.EqualMatch = func(ctx context.Context, want any, got any, ok bool) {
			fn(ctx, want, got, ok)
			extra.EqualMatch(ctx, want, got, ok)
		}
	}
	if extra.MatchTimeout != nil {
		fn := pt.MatchTimeout
		pt.MatchTimeout = func(ctx context.Context) {
			fn(ctx)
			extra.MatchTimeout(ctx)
		}
	}
	return pt
}

type EventInfo struct{}
