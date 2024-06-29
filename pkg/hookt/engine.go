package hookt

import (
	"context"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
	"golang.org/x/sync/errgroup"
	"hookt.dev/cmd/pkg/check"
	"hookt.dev/cmd/pkg/errors"
	"hookt.dev/cmd/pkg/plugin/builtin"
	"hookt.dev/cmd/pkg/proto"
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

func (e *Engine) Run(ctx context.Context, file string) (*check.S, error) {
	p, err := os.ReadFile(file)
	if err != nil {
		return nil, errors.New("failed to read file: %w", err)
	}

	w, err := e.p.Parse(ctx, p)
	if err != nil {
		return nil, errors.New("failed to parse file: %w", err)
	}

	var s check.S

	var g errgroup.Group

	for _, job := range w.Jobs {
		for _, step := range job.Steps {
			g.Go(func() error {
				r, ok := step.With.(proto.Runner)
				if !ok {
					return errors.New("step %q does not implement proto.Runner", step.ID)
				}

				defer r.Stop()

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

	return &s, g.Wait()
}
