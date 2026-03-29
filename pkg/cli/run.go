package cli

import (
	"context"
	"fmt"

	"github.com/suzuki-shunsuke/slog-util/slogutil"
	"github.com/suzuki-shunsuke/yamledit/pkg/cache"
	"github.com/suzuki-shunsuke/yamledit/pkg/controller"
	gh "github.com/suzuki-shunsuke/yamledit/pkg/github"
	"github.com/urfave/cli/v3"
)

func NewRun(logger *slogutil.Logger, gFlags *Flags) *cli.Command {
	var args []string
	return &cli.Command{
		Name:  "run",
		Usage: "Edit YAML files based on migration rules",
		Action: func(ctx context.Context, _ *cli.Command) error {
			if err := logger.SetLevel(gFlags.LogLevel); err != nil {
				return fmt.Errorf("set log level: %w", err)
			}
			migrations, yamlFiles, err := parseArgs(args)
			if err != nil {
				return fmt.Errorf("parse arguments: %w", err)
			}
			ghtknEnabled, err := gh.GetGHTKNEnabledFromEnv()
			if err != nil {
				return fmt.Errorf("get ghtkn enabled: %w", err)
			}
			ghClient := gh.New(ctx, logger.Logger, gh.GetGitHubTokenFromEnv(), ghtknEnabled)
			c := cache.New(gFlags.NoCache)
			return controller.Run(ctx, logger, ghClient, c, ".", migrations, yamlFiles)
		},
		Arguments: []cli.Argument{
			&cli.StringArgs{
				Name:        "args",
				Destination: &args,
				Min:         0,
				Max:         -1,
			},
		},
	}
}
