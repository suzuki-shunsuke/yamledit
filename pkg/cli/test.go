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

func NewTest(logger *slogutil.Logger, gFlags *Flags) *cli.Command {
	var migrations []string
	return &cli.Command{
		Name:  "test",
		Usage: "Test migration rules against expected results",
		Description: `Tests whether YAML files are modified as expected by ruleset files.
This command fails if any test files are not modified expectedly.

Rulesets are located at .yamledit/<ruleset>/ruleset.yaml.
Each test case is a subdirectory of .yamledit/<ruleset>/ containing test.yaml and result.yaml.
The ruleset is applied to test.yaml, and the result is verified against result.yaml.
This command doesn't change any files.
During testing, .rules[].files in ruleset files is ignored.
If a test directory has test.yaml without result.yaml, a warning is output and the test is skipped.
If no arguments are specified, all rulesets under .yamledit are tested. If no test directories exist, a warning is output and the ruleset is skipped.
If test files aren't modified expectedly, diffs are outputted.`,
		Action: func(ctx context.Context, _ *cli.Command) error {
			if err := logger.SetLevel(gFlags.LogLevel); err != nil {
				return fmt.Errorf("set log level: %w", err)
			}
			ghtknEnabled, err := gh.GetGHTKNEnabledFromEnv()
			if err != nil {
				return fmt.Errorf("get ghtkn enabled: %w", err)
			}
			ghClient := gh.New(ctx, logger.Logger, gh.GetGitHubTokenFromEnv(), ghtknEnabled)
			c := cache.New(gFlags.NoCache)
			return controller.Test(ctx, logger, ghClient, c, ".", migrations)
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
