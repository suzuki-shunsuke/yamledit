package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/suzuki-shunsuke/slog-util/slogutil"
	"github.com/suzuki-shunsuke/yamledit/pkg/config"
	"github.com/urfave/cli/v3"
)

// RunArgs holds the flag and argument values for the init command.
type RunArgs struct {
	*Flags
	Name string // positional argument
}

// NewInit creates a new init command instance with the provided logger.
// It returns a CLI command that can be registered with the main CLI application.
func NewInit(logger *slogutil.Logger, gFlags *Flags) *cli.Command {
	args := &RunArgs{
		Flags: gFlags,
	}
	return &cli.Command{
		Name:  "new",
		Usage: "Create a migration file using a default template if it doesn't exist",
		Description: `Create a migration file using a default template if it doesn't exist.
If the migration file already exists, this command does nothing.
The migration name must match the regular expression '^[a-z0-9_-]+$'.
The migration file is created at '.yamledit/<migration name>.yaml'.`,
		Action: func(ctx context.Context, _ *cli.Command) error {
			return runAction(ctx, logger, args)
		},
		Arguments: []cli.Argument{
			&cli.StringArg{
				Name:        "<migration name>",
				Destination: &args.Name,
			},
		},
	}
}

func runAction(_ context.Context, logger *slogutil.Logger, args *RunArgs) error {
	if err := logger.SetLevel(args.LogLevel); err != nil {
		return fmt.Errorf("set log level: %w", err)
	}
	if args.Name == "" {
		return errors.New("migration name is required")
	}

	if err := config.New(".", args.Name); err != nil {
		return fmt.Errorf("initialize config: %w", err)
	}
	return nil
}
