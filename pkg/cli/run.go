package cli

import (
	"context"
	"fmt"

	"github.com/suzuki-shunsuke/slog-util/slogutil"
	"github.com/suzuki-shunsuke/yamledit/pkg/controller"
	gh "github.com/suzuki-shunsuke/yamledit/pkg/github"
	"github.com/urfave/cli/v3"
)

func NewRun(logger *slogutil.Logger, gFlags *Flags) *cli.Command {
	var yamlFiles []string
	return &cli.Command{
		Name:  "run",
		Usage: "Edit YAML files based on migration rules",
		Action: func(ctx context.Context, _ *cli.Command) error {
			if err := logger.SetLevel(gFlags.LogLevel); err != nil {
				return fmt.Errorf("set log level: %w", err)
			}
			ghtknEnabled, err := gh.GetGHTKNEnabledFromEnv()
			if err != nil {
				return fmt.Errorf("get ghtkn enabled: %w", err)
			}
			ghClient := gh.New(ctx, logger.Logger, gh.GetGitHubTokenFromEnv(), ghtknEnabled)
			return controller.Run(ctx, logger, ghClient, ".", yamlFiles)
		},
		Arguments: []cli.Argument{
			&cli.StringArgs{
				Name:        "yaml files",
				Destination: &yamlFiles,
				Min:         1,
				Max:         -1,
			},
		},
	}
}
