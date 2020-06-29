package cmd

import (
	"github.com/spf13/cobra"

	"gerrit.instructure.com/muss/config"
)

func newRecreateCommand(cfg *config.ProjectConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "recreate",
		Short: "Recreate containers",
		Long: `Stop, rm, and recreate running containers to apply latest changes.

This is a shortcut for "up --detach --force-recreate --renew-anon-volumes".

This is useful for truncating container logs and refreshing environment settings.`,
		Args: cobra.ArbitraryArgs,
		// TODO: ArgsInUseLine: "[service...]"
		PreRunE: configSavePreRun(cfg),
		RunE: func(cmd *cobra.Command, args []string) error {
			args = append([]string{"--detach", "--force-recreate", "--renew-anon-volumes"}, args...)

			return cmdDelegator(cmd).Delegate(
				dockerComposeNamedCmd("up", cmd, args),
			)
		},
	}

	cmd.Flags().BoolP("no-build", "", false, "Don't build an image, even if it's missing.")
	cmd.Flags().BoolP("no-start", "", false, "Don't start the services after creating them.")
	cmd.Flags().BoolP("build", "", false, "Build images before starting containers.")
	cmd.Flags().IntP("timeout", "t", 10, "Use this `timeout` in seconds for container\nshutdown when attached or when containers are\nalready running. (default: 10)")
	cmd.Flags().StringP("scale", "", "", "With `SERVICE=NUM` scale SERVICE to NUM instances.\nOverrides the `scale` setting in the Compose file if present.")

	return cmd
}

func init() {
	AddCommandBuilder(newRecreateCommand)
}
