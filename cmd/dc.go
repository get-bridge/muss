package cmd

import (
	"github.com/spf13/cobra"

	"github.com/instructure-bridge/muss/config"
	"github.com/instructure-bridge/muss/proc"
)

func newDcCommand(cfg *config.ProjectConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "dc",
		Short: "Call aribtrary docker-compose commands",
		Long: `Shortcut for calling any docker-compose commands directly
after processing config.

Useful to run anything muss doesn't have a native command for
or if you need to work around the way muss wraps docker-compose.

Usage:
  muss dc events --json
  muss dc config --resolve-image-digests
  muss dc images
`,
		Args:               cobra.ArbitraryArgs,
		DisableFlagParsing: true,
		PreRunE:            configSavePreRun(cfg),
		RunE: func(cmd *cobra.Command, args []string) error {

			return proc.Exec(append([]string{"docker-compose"}, args...))
		},
	}

	// Just show "Long" by providing a zero-width (but not zero-value) string.
	cmd.SetUsageTemplate(`{{ "" }}`)

	return cmd
}

func init() {
	AddCommandBuilder(newDcCommand)
}
