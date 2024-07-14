package hookt

import (
	"context"
	"log/slog"
	"strconv"

	"hookt.dev/cmd/pkg/check"
	"hookt.dev/cmd/pkg/errors"
	"hookt.dev/cmd/pkg/plugin/builtin"
	"hookt.dev/cmd/pkg/proto"
	"hookt.dev/cmd/pkg/trace"

	"github.com/lmittmann/tint"
	"golang.org/x/sync/errgroup"
)

var plugins []proto.Interface

func init() {
	for _, p := range builtin.Plugins() {
		plugins = append(plugins, p)
	}
}

type Engine struct {
	p *proto.P
}

func New(opts ...func(*Engine)) *Engine {
	ngn := &Engine{
		p: proto.New(
			proto.WithPlugins(plugins...),
		),
	}
	for _, opt := range opts {
		opt(ngn)
	}
	return ngn
}

func (e *Engine) Run(ctx context.Context, p []byte) (*check.S, error) {
	var s check.S

	ctx = trace.WithPattern(ctx, trace.ContextPattern(ctx).Join(s.Trace()))

	w, err := e.p.Parse(ctx, p)
	if err != nil {
		return nil, errors.New("failed to parse file: %w", err)
	}

	var g errgroup.Group

	for i, job := range w.Jobs {
		ctx := trace.With(ctx, "job", job.ID)
		ctx = trace.With(ctx, "job-index", strconv.Itoa(i))
		for j, step := range job.Steps {
			ctx := trace.With(ctx, "step", step.ID)
			ctx = trace.With(ctx, "step-desc", step.Desc)
			ctx = trace.With(ctx, "step-index", strconv.Itoa(j))
			g.Go(func() error {
				r, ok := step.With.(proto.Runner)
				if !ok {
					return errors.New("step %q does not implement proto.Runner", step.ID)
				}

				defer r.Stop(ctx)

				if err := r.Run(ctx, &s); err != nil {
					slog.Error("step failure",
						"desc", step.Desc,
						tint.Err(err),
					)

					return err
				}

				slog.Info("step pass",
					"desc", step.Desc,
				)

				return nil
			})
		}
	}

	done := make(chan error)

	go func() {
		done <- g.Wait()
	}()

	select {
	case <-ctx.Done():
		return &s, ctx.Err()
	case err := <-done:
		return &s, err
	}
}
