package command // import "hookt.dev/cmd/pkg/command"

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"hookt.dev/cmd/pkg/errors"
	"hookt.dev/cmd/pkg/hookt"

	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"sigs.k8s.io/yaml"
)

type BuildInfo struct {
	Version string
	Commit  string
	Date    string
}

type App struct {
	BuildInfo

	Engine *hookt.Engine

	debug bool
}

func (app *App) Register(f *pflag.FlagSet) {
	f.BoolVar(&app.debug, "debug", app.debug, "enable debug logging")
}

func New(name string, opts ...func(*App)) *App {
	app := &App{
		Engine: hookt.New(),
	}

	for _, opt := range opts {
		opt(app)
	}

	return app
}

func (app *App) Init(cmd *cobra.Command, args []string) {
	level := slog.LevelInfo
	if app.debug {
		level = slog.LevelDebug
	}

	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			AddSource:  true,
			Level:      level,
			TimeFormat: time.Kitchen,
		}),
	))
}

func (app *App) Render(v any) error {
	p, err := yaml.Marshal(v)
	if err != nil {
		return errors.New("rendering failed: %w", err)
	}
	fmt.Printf("%s", p)
	return nil
}
