package proto

import (
	"bytes"
	"context"
	"math/rand"
	"os"
	"strings"
	"text/template"

	"hookt.dev/cmd/pkg/async"
	"hookt.dev/cmd/pkg/errors"

	"github.com/Masterminds/sprig"
	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/yaml"
)

type TOption func(*template.Template) *template.Template

type T struct {
	Options []TOption
	Vars    async.Map
}

func (p *P) Evaluate(tmpl string, data any) ([]byte, error) {
	var buf bytes.Buffer

	t, err := p.t.Parse("", tmpl)
	if err != nil {
		return nil, errors.New("failed to parse template %q: %w", tmpl, err)
	}

	if err := t.Execute(&buf, data); err != nil {
		return nil, errors.New("failed to evaluate template %q: %w", tmpl, err)
	}

	return buf.Bytes(), nil
}

func (t *T) Parse(name, data string) (*template.Template, error) {
	tmpl := template.New(name).
		Funcs(sprig.FuncMap()).
		Funcs(t.funcs()).
		Delims("${{", "}}").
		Option("missingkey=error")

	for _, opt := range t.Options {
		tmpl = opt(tmpl)
	}

	return tmpl.Parse(data)
}

func (t *T) funcs() template.FuncMap {
	return map[string]any{
		"xrand":    xrand,
		"setvar":   t.set,
		"seterror": t.seterror,
		"var":      t.get,
		"setenv":   os.Setenv,
		"env":      os.Getenv,
	}
}

func (t *T) Match(ctx context.Context, data string) func(context.Context, any) (bool, error) {
	tmpl, err := t.Parse("", data)
	if err != nil {
		return func(context.Context, any) (bool, error) { return false, err }
	}
	return func(_ context.Context, x any) (bool, error) {
		var buf bytes.Buffer

		if err := tmpl.Execute(&buf, x); err != nil {
			return false, errors.New("failed to evaluate %q: %w", data, err)
		}

		var v any

		if err := yaml.Unmarshal(buf.Bytes(), &v); err != nil {
			return false, errors.New("failed to parse result: %w", err)
		}

		switch v := v.(type) {
		case bool:
			return v, nil
		default:
			return cmp.Equal(x, v), nil
		}
	}
}

func (t *T) set(name string, value any) any {
	t.Vars.Store(name, value)
	return value
}

func (t *T) get(name string) any {
	value, _ := t.Vars.Load(name)
	return value
}

func (t *T) seterror(err string) bool {
	// TODO: implement
	return true
}

func xrand(s string) string {
	var (
		buf strings.Builder
		n   int
		ok  bool
	)
	for _, r := range s {
		if r == 'X' {
			n, ok = n+1, true
		} else {
			if ok {
				buf.WriteString(gen(n))
				n, ok = 0, false
			}
			buf.WriteRune(r)
		}
	}
	if ok {
		buf.WriteString(gen(n))
	}
	return buf.String()
}

func gen(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	s := make([]byte, n)
	for i := range s {
		s[i] = charset[rand.Intn(len(charset))]
	}
	return string(s)
}
