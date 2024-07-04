package proto

import (
	"context"
	"fmt"
	"log/slog"
	"sort"

	"hookt.dev/cmd/pkg/errors"
	"hookt.dev/cmd/pkg/proto/wire"
	"hookt.dev/cmd/pkg/trace"

	"github.com/google/go-cmp/cmp"
	"github.com/itchyny/gojq"
	"sigs.k8s.io/yaml"
)

type Pattern struct {
	Key   *gojq.Query
	Match func(context.Context, any) (bool, error)
}

type Patterns []*Pattern

func (p Patterns) Match(ctx context.Context, obj any) (bool, error) {
	for _, p := range p {
		ctx := pattern(ctx, p.Key.String())

		it := p.Key.RunWithContext(ctx, obj)

		slog.Debug("pattern",
			"query", p.Key.String(),
		)

		// TODO: Handle multiple results?
		v, ok := it.Next()
		if !ok {
			return false, nil
		}

		ok, err := p.Match(ctx, v)
		if err != nil {
			return false, errors.New("failed to match jq %q: %w", p.Key.String(), err)
		}
		if !ok {
			return false, nil
		}
	}

	return true, nil
}

func pattern(ctx context.Context, name string) context.Context {
	return trace.With(ctx, "pattern", name)
}

func (p *P) Patterns(ctx context.Context, obj wire.Object, opts ...TOption) (Patterns, error) {
	var (
		pt  = make(Patterns, 0, len(obj))
		tr  = trace.ContextPattern(ctx)
		t   = p.t.With(opts...)
		err error
	)

	for k, raw := range obj {
		var (
			e error
			q Pattern
		)

		ctx := pattern(ctx, k)

		q.Key, e = gojq.Parse(k)
		tr.ParseKey(ctx, q.Key, e)
		if e != nil {
			err = errors.Join(
				err,
				errors.New("failed to parse jq %q: %w", k, e),
			)
			continue
		}

		var want any

		e = yaml.Unmarshal(raw, &want)
		tr.UnmarshalValue(ctx, raw, want, e)
		if e != nil {
			err = errors.Join(
				err,
				errors.New("failed to parse value for jq %q: %w", k, e),
			)
			continue
		}

		slog.Debug("building pattern",
			"key", k,
			"pattern", want,
		)

		switch want := want.(type) {
		case bool:
			q.Match = func(_ context.Context, got any) (bool, error) {
				ok := want && len(raw) != 0
				tr.EqualMatch(ctx, want, got, ok)
				return ok, nil
			}
		case string:
			q.Match = t.Match(ctx, want)
		default:
			q.Match = func(_ context.Context, got any) (bool, error) {
				ok := cmpEqual(want, got)
				tr.EqualMatch(ctx, want, got, ok)
				return ok, nil
			}
		}

		pt = append(pt, &q)
	}

	sort.Slice(pt, func(i, j int) bool {
		return pt[i].Key.String() < pt[j].Key.String()
	})

	return pt, err
}

func cmpEqual(want, got any) bool {
	if fmt.Sprint(want) == fmt.Sprint(got) {
		return true
	}
	return cmp.Equal(want, got)
}
