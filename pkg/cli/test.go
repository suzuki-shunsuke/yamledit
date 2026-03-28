package cli

import (
	"context"
	"fmt"

	"github.com/suzuki-shunsuke/slog-util/slogutil"
	"github.com/suzuki-shunsuke/yamledit/pkg/controller"
	"github.com/urfave/cli/v3"
)

func NewTest(logger *slogutil.Logger, gFlags *Flags) *cli.Command {
	var migrations []string
	return &cli.Command{
		Name:  "test",
		Usage: "Test migration rules against expected results",
		Action: func(ctx context.Context, _ *cli.Command) error {
			if err := logger.SetLevel(gFlags.LogLevel); err != nil {
				return fmt.Errorf("set log level: %w", err)
			}
			return controller.Test(ctx, logger, ".", migrations)
		},
		Arguments: []cli.Argument{
			&cli.StringArgs{
				Name:        "migrations",
				Destination: &migrations,
				Min:         0,
				Max:         -1,
			},
		},
	}
}
