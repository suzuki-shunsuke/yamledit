// Package cli provides the command-line interface layer for yamledit.
// This package serves as the main entry point for all CLI operations,
// handling command parsing, flag processing, and routing to appropriate subcommands.
// It orchestrates the overall CLI structure using urfave/cli framework and delegates
// actual business logic to controller packages.
package cli

import (
	"context"

	"github.com/suzuki-shunsuke/slog-util/slogutil"
	"github.com/suzuki-shunsuke/urfave-cli-v3-util/urfave"
	"github.com/urfave/cli/v3"
)

func Run(ctx context.Context, logger *slogutil.Logger, env *urfave.Env) error {
	gFlags := &Flags{}
	return urfave.Command(env, &cli.Command{ //nolint:wrapcheck
		Name:  "yamledit",
		Usage: "Edit YAML files based on declarative rules. https://github.com/suzuki-shunsuke/yamledit",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "log-level",
				Usage:       "Log level (debug, info, warn, error)",
				Sources:     cli.EnvVars("YAMLEDIT_LOG_LEVEL"),
				Destination: &gFlags.LogLevel,
				Local:       true,
			},
			&cli.BoolFlag{
				Name:        "no-cache",
				Usage:       "Ignore cache and always fetch remote imports",
				Destination: &gFlags.NoCache,
				Local:       true,
			},
		},
		Commands: []*cli.Command{
			NewInit(logger, gFlags),
			NewRun(logger, gFlags),
			NewTest(logger, gFlags),
			NewAdd(logger, gFlags),
			NewSearch(logger, gFlags),
		},
	}).Run(ctx, env.Args)
}
