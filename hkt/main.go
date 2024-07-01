package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"hookt.dev/cmd/pkg/command"
	"hookt.dev/cmd/pkg/trace"

	"github.com/spf13/cobra"
)

var (
	version string
	commit  string
	date    string
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	var (
		app = command.New("hkt",
			command.WithBuildInfo(version, commit, date),
		)
		cmd = newCommand(ctx, app)
	)

	app.Register(cmd.PersistentFlags())

	if err := cmd.Execute(); err != nil {
		die(err)
	}
}

func die(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func newCommand(ctx context.Context, app *command.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:              "hkt",
		Short:            "CLI for testing",
		Args:             cobra.NoArgs,
		PersistentPreRun: app.Init,
		Version:          version,
	}

	cmd.AddCommand(
		newRunCommand(ctx, app),
	)

	return cmd
}

func newRunCommand(ctx context.Context, app *command.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "run",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, files []string) error {
			if cmd.Flags().Changed("debug") {
				ctx = trace.WithJob(ctx, trace.LogJob())
				ctx = trace.WithPattern(ctx, trace.LogPattern())
				ctx = trace.WithSchedule(ctx, trace.LogSchedule())
			}

			s, err := app.Engine.Run(ctx, files[0])
			if s != nil && len(s.Events) != 0 {
				app.Render(s.Results())
			}
			return err
		},
		Version:      version,
		SilenceUsage: true,
	}

	return cmd
}
