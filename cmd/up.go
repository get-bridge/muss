package cmd

import (
	"github.com/spf13/cobra"

	"gerrit.instructure.com/muss/config"
)

func newUpCommand() *cobra.Command {
	opts := struct {
		detach               bool
		noColor              bool
		quietPull            bool
		noDeps               bool
		forceRecreate        bool
		alwaysRecreateDeps   bool
		noRecreate           bool
		noBuild              bool
		noStart              bool
		build                bool
		abortOnContainerExit bool
		timeout              int
		renewAnonVolumes     bool
		removeOrphans        bool
		exitCodeFrom         string
		scale                string
	}{}

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
`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			config.Save()

			return DelegateCmd(
				cmd,
				dockerComposeCmd(cmd, args),
			)
		},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().BoolVarP(&opts.detach, "detach", "d", false, "Detached mode: Run containers in the background,\nprint new container names. Incompatible with\n--abort-on-container-exit.")
	cmd.Flags().BoolVarP(&opts.noColor, "no-color", "", false, "Produce monochrome output.")
	cmd.Flags().BoolVarP(&opts.quietPull, "quiet-pull", "", false, "Pull without printing progress information")
	cmd.Flags().BoolVarP(&opts.noDeps, "no-deps", "", false, "Don't start linked services.")
	cmd.Flags().BoolVarP(&opts.forceRecreate, "force-recreate", "", false, "Recreate containers even if their configuration\nand image haven't changed.")
	cmd.Flags().BoolVarP(&opts.alwaysRecreateDeps, "always-recreate-deps", "", false, "Recreate dependent containers.\nIncompatible with --no-recreate.")
	cmd.Flags().BoolVarP(&opts.noRecreate, "no-recreate", "", false, "If containers already exist, don't recreate\nthem. Incompatible with --force-recreate and -V.")
	cmd.Flags().BoolVarP(&opts.noBuild, "no-build", "", false, "Don't build an image, even if it's missing.")
	cmd.Flags().BoolVarP(&opts.noStart, "no-start", "", false, "Don't start the services after creating them.")
	cmd.Flags().BoolVarP(&opts.build, "build", "", false, "Build images before starting containers.")
	cmd.Flags().BoolVarP(&opts.abortOnContainerExit, "abort-on-container-exit", "", false, "Stops all containers if any container was\nstopped. Incompatible with -d.")
	cmd.Flags().IntVarP(&opts.timeout, "timeout", "t", 10, "Use this `timeout` in seconds for container\nshutdown when attached or when containers are\nalready running. (default: 10)")
	cmd.Flags().BoolVarP(&opts.renewAnonVolumes, "renew-anon-volumes", "V", false, "Recreate anonymous volumes instead of retrieving\ndata from the previous containers.")
	cmd.Flags().BoolVarP(&opts.removeOrphans, "remove-orphans", "", false, "Remove containers for services not defined\nin the Compose file.")
	cmd.Flags().StringVarP(&opts.exitCodeFrom, "exit-code-from", "", "", "Return the exit code of the selected `service`\ncontainer. Implies --abort-on-container-exit.")
	cmd.Flags().StringVarP(&opts.scale, "scale", "", "", "With `SERVICE=NUM` scale SERVICE to NUM instances.\nOverrides the `scale` setting in the Compose file if present.")

	return cmd
}

func init() {
	rootCmd.AddCommand(newUpCommand())
}
