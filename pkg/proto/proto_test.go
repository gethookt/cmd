package proto_test

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"hookt.dev/cmd/pkg/plugin/builtin"
	"hookt.dev/cmd/pkg/proto"

	"github.com/lmittmann/tint"
)

func init() {
	slog.SetDefault(
		slog.New(
			tint.NewHandler(os.Stderr, &tint.Options{
				AddSource:  true,
				Level:      slog.LevelDebug,
				TimeFormat: time.Kitchen,
			}),
		),
	)
}

func newP() *proto.P {
	p := builtin.Plugins()
	q := make([]proto.Interface, len(p))
	for i, p := range p {
		q[i] = p
	}
	return proto.New(
		proto.WithPlugins(q...),
	)
}

func TestParse(t *testing.T) {
	p := newP()
	q := file(t, "../testdata/ok.yaml")
	ctx := context.Background()

	w, err := p.Parse(ctx, q)
	if err != nil {
		t.Fatal(err)
	}

	_ = w

	// spew.Dump(w)
}

func file(t *testing.T, path string) []byte {
	t.Helper()

	p, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	return p
}
