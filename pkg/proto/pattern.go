package proto

import (
	"context"
	"fmt"
	"log/slog"
	"text/template"

	"hookt.dev/cmd/pkg/errors"
	"hookt.dev/cmd/pkg/proto/wire"
	"hookt.dev/cmd/pkg/trace"

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
	var (
		pt  = make(Pattern)
		tr  = trace.ContextPattern(ctx)
		err error
	)

	for k, raw := range obj {
		q, e := gojq.Parse(k)
		tr.ParseKey(k, q, e)
		if e != nil {
			err = errors.Join(
				err,
				errors.New("failed to parse jq %q: %w", k, e),
			)
			continue
		}

		var want any

		e = yaml.Unmarshal(raw, &want)
		tr.UnmarshalValue(k, raw, want, e)
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
			pt[q] = func(_ context.Context, got any) (bool, error) { return want || got == nil, nil }
		case string:
			tv := tr.TemplateValue
			tr.TemplateValue = func(_, v string, t *template.Template, e error) { tv(k, v, t, e) }
			pt[q] = p.t.Match(trace.WithPattern(ctx, tr), want)
		default:
			pt[q] = func(_ context.Context, got any) (bool, error) {
				ok := cmpEqual(want, got)
				tr.EqualMatch(k, want, got, ok)
				return ok, nil
			}
		}
	}

	return pt, err
}

func cmpEqual(want, got any) bool {
	if fmt.Sprint(want) == fmt.Sprint(got) {
		return true
	}
	return cmp.Equal(want, got)
}
