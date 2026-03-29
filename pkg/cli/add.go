package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/suzuki-shunsuke/slog-util/slogutil"
	"github.com/suzuki-shunsuke/yamledit/pkg/cache"
	"github.com/suzuki-shunsuke/yamledit/pkg/controller"
	gh "github.com/suzuki-shunsuke/yamledit/pkg/github"
	"github.com/urfave/cli/v3"
)

type AddArgs struct {
	*Flags
	Alias     string
	Migration string
	Force     bool
}

func NewAdd(logger *slogutil.Logger, gFlags *Flags) *cli.Command {
	args := &AddArgs{
		Flags: gFlags,
	}
	return &cli.Command{
		Name:  "add",
		Usage: "Add a Reusable Rule alias to the configuration file",
		Action: func(ctx context.Context, _ *cli.Command) error {
			return addAction(ctx, logger, args)
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "force",
				Aliases:     []string{"f"},
				Usage:       "Overwrite existing entry",
				Destination: &args.Force,
			},
		},
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:        "<alias>",
				Destination: &args.Alias,
			},
			&cli.StringArg{
				Name:        "<migration>",
				Destination: &args.Migration,
			},
		},
	}
}

func addAction(ctx context.Context, logger *slogutil.Logger, args *AddArgs) error {
	if err := logger.SetLevel(args.LogLevel); err != nil {
		return fmt.Errorf("set log level: %w", err)
	}
	if args.Alias == "" {
		return errors.New("alias is required")
	}
	if args.Migration == "" {
		return errors.New("migration is required")
	}
	ghtknEnabled, err := gh.GetGHTKNEnabledFromEnv()
	if err != nil {
		return fmt.Errorf("get ghtkn enabled: %w", err)
	}
	ghClient := gh.New(ctx, logger.Logger, gh.GetGitHubTokenFromEnv(), ghtknEnabled)
	c := cache.New(args.NoCache)
	return controller.Add(ctx, logger, ghClient, c, ".", args.Alias, args.Migration, args.Force) //nolint:wrapcheck
}
