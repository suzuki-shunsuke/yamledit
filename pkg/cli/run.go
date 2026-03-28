package cli

import (
	"context"
	"fmt"

	"github.com/suzuki-shunsuke/slog-util/slogutil"
	"github.com/suzuki-shunsuke/yamledit/pkg/controller"
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
			return controller.Run(ctx, logger, ".", yamlFiles)
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
