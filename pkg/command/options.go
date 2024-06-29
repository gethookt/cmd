package command

import (
	"hookt.dev/cmd/pkg/hookt"
)

func WithBuildInfo(version, commit, date string) func(app *App) {
	return func(app *App) {
		app.BuildInfo = BuildInfo{
			Version: version,
			Commit:  commit,
			Date:    date,
		}
	}
}

func WithEngineOptions(opts ...func(*hookt.Engine)) func(app *App) {
	return func(app *App) {
		app.Engine = hookt.New(opts...)
	}
}
