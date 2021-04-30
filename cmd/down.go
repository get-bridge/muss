package cmd

import (
	"github.com/spf13/cobra"

	"github.com/get-bridge/muss/config"
)

func newDownCommand(cfg *config.ProjectConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "down",
		Short: "Stop and remove containers, networks, images, and volumes",
		Long: `Stops containers and removes containers, networks, volumes, and images
created by "up".

By default, the only things removed are:

- Containers for services
- Networks defined in the "networks" section
- The default network, if one is used

Networks and volumes defined as "external" are never removed.

Usage: down [options]

Options:
  --rmi type              Remove images. Type must be one of:
                            'all': Remove all images used by any service.
                            'local': Remove only images that don't have a
                            custom tag set by the "image" field.
  -v, --volumes           Remove named volumes declared in the "volumes"
                          section of the Compose file and anonymous volumes
                          attached to containers.
  --remove-orphans        Remove containers for services not defined in the
                          Compose file
  -t, --timeout TIMEOUT   Specify a shutdown timeout in seconds.
                          (default: 10)
`,
		Args:               cobra.ArbitraryArgs,
		DisableFlagParsing: true,
		PreRunE:            configSavePreRun(cfg),
		RunE: func(cmd *cobra.Command, args []string) error {
			return DelegateCmd(
				cmd,
				dockerComposeCmd(cmd, args),
			)
		},
	}

	// Just show "Long" by providing a zero-width but not zero-value string.
	cmd.SetUsageTemplate(`{{ "" }}`)

	return cmd
}

func init() {
	AddCommandBuilder(newDownCommand)
}
