package cmd

import (
	"github.com/spf13/cobra"

	"gerrit.instructure.com/muss/config"
)

func newRestartCommand(cfg *config.ProjectConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "restart",
		Short: "Restart services",
		Long:  "Restart running containers.",
		Args:  cobra.ArbitraryArgs,
		// TODO: ArgsInUseLine: "[service...]"
		PreRun: configSavePreRun,
		RunE: func(cmd *cobra.Command, args []string) error {
			return DelegateCmd(
				cmd,
				dockerComposeCmd(cmd, args),
			)
		},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().IntP("timeout", "t", 10, "Specify a shutdown `timeout` in seconds.")

	return cmd
}

func init() {
	AddCommandBuilder(newRestartCommand)
}
