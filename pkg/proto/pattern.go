package proto

import (
	"context"
	"log/slog"

	"hookt.dev/cmd/pkg/errors"
	"hookt.dev/cmd/pkg/proto/wire"

	"github.com/google/go-cmp/cmp"
	"github.com/itchyny/gojq"
	"sigs.k8s.io/yaml"
)

type Pattern map[*gojq.Query]func(context.Context, any) (bool, error)

func (p Pattern) Match(ctx context.Context, obj any) (bool, error) {
	for q, fn := range p {
		it := q.RunWithContext(ctx, obj)

		slog.Debug("pattern",
			"query", q.String(),
		)

		// TODO: Handle multiple results?
		v, ok := it.Next()
		if !ok {
			return false, nil
		}

		ok, err := fn(ctx, v)
		if err != nil {
			return false, errors.New("failed to match jq %q: %w", q.String(), err)
		}
		if !ok {
			return false, nil
		}
	}

	return len(p) != 0, nil
}

func (p *P) Pattern(ctx context.Context, obj wire.Object) (Pattern, error) {
	pt := make(Pattern)

	for k, raw := range obj {
		q, err := gojq.Parse(k)
		if err != nil {
			return nil, errors.New("failed to parse jq %q: %w", k, err)
		}

		var v any

		if err := yaml.Unmarshal(raw, &v); err != nil {
			return nil, errors.New("failed to parse value for jq %q: %w", k, err)
		}

		slog.Debug("building pattern",
			"key", k,
			"pattern", v,
		)

		switch v := v.(type) {
		case bool:
			pt[q] = func(_ context.Context, x any) (bool, error) { return v || x == nil, nil }
		case string:
			pt[q] = p.t.Match(ctx, v)
		default:
			pt[q] = func(_ context.Context, x any) (bool, error) { return cmp.Equal(x, v), nil }
		}
	}

	return pt, nil
}
