package proto_test

import (
	"context"
	"testing"

	"hookt.dev/cmd/pkg/proto/wire"
)

func TestPattern(t *testing.T) {
	cases := []struct {
		raw wire.Object
		obj any
		ok  bool
	}{
		0: {
			wire.Object{
				".foo.one": []byte(`"bar"`),
			},
			map[string]any{
				"foo": map[string]any{
					"one": "bar",
				},
			},
			true,
		},
		1: {
			wire.Object{
				".foo.two": []byte(`"10"`),
			},
			map[string]any{
				"foo": map[string]any{
					"two": "10",
				},
			},
			true,
		},
		2: {
			wire.Object{
				".foo.three": []byte(`true`),
			},
			map[string]any{
				"foo": map[string]any{
					"three": "123",
				},
			},
			true,
		},
		3: {
			wire.Object{
				".foo.one":   []byte(`"bar"`),
				".foo.two":   []byte(`"10"`),
				".foo.three": []byte(`true`),
			},
			map[string]any{
				"foo": map[string]any{
					"one":   "bar",
					"two":   "10",
					"three": "123",
				},
			},
			true,
		},
		4: {
			wire.Object{
				".foo[0]": []byte(`"1"`),
				".foo[1]": []byte(`"2"`),
				".foo[2]": []byte(`"3"`),
			},
			map[string]any{
				"foo": []any{"1", "2", "3"},
			},
			true,
		},
		5: {
			wire.Object{
				".foo.bar": []byte(`"${{ setvar "bar" . }}"`),
			},
			map[string]any{
				"foo": map[string]any{
					"bar": "magic",
				},
			},
			true,
		},
		6: {
			wire.Object{
				".foo.bar": []byte(`"${{ var "bar" }}"`),
			},
			map[string]any{
				"foo": map[string]any{
					"bar": "magic",
				},
			},
			true,
		},
		7: {
			wire.Object{
				".foo.two": []byte(`false`),
			},
			map[string]any{
				"foo": map[string]any{
					"one": "rab",
				},
			},
			true,
		},

		8: {
			wire.Object{
				".foo.one": []byte(`"bar"`),
			},
			map[string]any{
				"foo": map[string]any{
					"one": "rab",
				},
			},
			false,
		},
		9: {
			wire.Object{
				".foo.one":   []byte(`"bar"`),
				".foo.two":   []byte(`"10"`),
				".foo.three": []byte(`false`),
			},
			map[string]any{
				"foo": map[string]any{
					"one":   "bar",
					"two":   "10",
					"three": "123",
				},
			},
			false,
		},
	}

	p := newP()
	ctx := context.Background()

	for _, cas := range cases {
		t.Run("", func(t *testing.T) {
			pt, err := p.Pattern(ctx, cas.raw)
			if err != nil {
				t.Fatal(err)
			}

			ok, err := pt.Match(ctx, cas.obj)
			if err != nil {
				t.Fatal(err)
			}

			if ok != cas.ok {
				t.Errorf("match: got %t, want %t", ok, cas.ok)
			}
		})
	}
}
