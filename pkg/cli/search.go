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

type SearchArgs struct {
	*Flags
	args []string
}

func NewSearch(logger *slogutil.Logger, gFlags *Flags) *cli.Command {
	args := &SearchArgs{
		Flags: gFlags,
	}
	return &cli.Command{
		Name:  "search",
		Usage: "Search for reusable rules",
		Action: func(ctx context.Context, _ *cli.Command) error {
			var query string
			if len(args.args) > 0 {
				query = args.args[0]
			}
			return searchAction(ctx, logger, args, query)
		},
		Arguments: []cli.Argument{
			&cli.StringArgs{
				Name:        "query",
				Destination: &args.args,
				Min:         0,
				Max:         1,
			},
		},
	}
}

func searchAction(ctx context.Context, logger *slogutil.Logger, args *SearchArgs, query string) error {
	if err := logger.SetLevel(args.LogLevel); err != nil {
		return fmt.Errorf("set log level: %w", err)
	}
	ghtknEnabled, err := gh.GetGHTKNEnabledFromEnv()
	if err != nil {
		return fmt.Errorf("get ghtkn enabled: %w", err)
	}
	ghClient := gh.New(ctx, logger.Logger, gh.GetGitHubTokenFromEnv(), ghtknEnabled)
	c := cache.New(args.NoCache)
	return controller.Search(ctx, logger, ghClient, c, query) //nolint:wrapcheck
}
