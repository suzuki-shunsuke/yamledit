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
		Description: `Tests whether YAML files are modified as expected by migration files.
This command fails if any test files are not modified expectedly.

Tests for .yamledit/<migration>.yaml are located at .yamledit/<migration>_test/.
The migration is applied to .yamledit/<migration>_test/<test>.yaml, and the result is verified against .yamledit/<migration>_test/<test>_result.yaml.
This command doesn't change any files.
During testing, .rules[].files in migration files is ignored.
Files ending with _result.yaml are excluded from test targets.
If a <test>.yaml exists without a corresponding <test>_result.yaml, a warning is output and <test>.yaml is skipped.
If no arguments are specified, all migration files under .yamledit are tested. If no test files exist, a warning is output and the migration is skipped.
If test files aren't modified expectedly, diff are outputted.`,
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
