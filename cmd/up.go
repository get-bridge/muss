package cmd

import (
	"github.com/spf13/cobra"

	"gerrit.instructure.com/muss/config"
)

func newUpCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "up",
		Short: "Create and start containers",
		Long: `Builds, (re)creates, starts, and attaches to containers for a service.

Unless they are already running, this command also starts any linked services.

The "up" command aggregates the output of each container. When
the command exits, all containers are stopped. Running "up -d"
starts the containers in the background and leaves them running.

If there are existing containers for a service, and the service's configuration
or image was changed after the container's creation, "up" picks
up the changes by stopping and recreating the containers (preserving mounted
volumes). To prevent picking up changes, use the "--no-recreate" flag.

If you want to force Compose to stop and recreate all containers, use the
"--force-recreate" flag.

Options:
  -d, --detach               Detached mode: Run containers in the background,
                             print new container names. Incompatible with
                             --abort-on-container-exit.
  --no-color                 Produce monochrome output.
  --quiet-pull               Pull without printing progress information
  --no-deps                  Don't start linked services.
  --force-recreate           Recreate containers even if their configuration
                             and image haven't changed.
  --always-recreate-deps     Recreate dependent containers.
                             Incompatible with --no-recreate.
  --no-recreate              If containers already exist, don't recreate
                             them. Incompatible with --force-recreate and -V.
  --no-build                 Don't build an image, even if it's missing.
  --no-start                 Don't start the services after creating them.
  --build                    Build images before starting containers.
  --abort-on-container-exit  Stops all containers if any container was
                             stopped. Incompatible with -d.
  -t, --timeout TIMEOUT      Use this timeout in seconds for container
                             shutdown when attached or when containers are
                             already running. (default: 10)
  -V, --renew-anon-volumes   Recreate anonymous volumes instead of retrieving
                             data from the previous containers.
  --remove-orphans           Remove containers for services not defined
                             in the Compose file.
  --exit-code-from SERVICE   Return the exit code of the selected service
                             container. Implies --abort-on-container-exit.
  --scale SERVICE=NUM        Scale SERVICE to NUM instances. Overrides the
                             "scale" setting in the Compose file if present.
`,
		Args:               cobra.ArbitraryArgs,
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			config.Save()

			return DelegateCmd(
				cmd,
				dockerComposeCmd(cmd, args),
			)
		},
	}

	// Just show "Long" by providing a zero-width (but not zero-value) string.
	cmd.SetUsageTemplate(`{{ "" }}`)

	return cmd
}

func init() {
	rootCmd.AddCommand(newUpCommand())
}
