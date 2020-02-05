package cmd

import (
	"github.com/spf13/cobra"

	"gerrit.instructure.com/muss/config"
)

func newStopCommand(cfg *config.ProjectConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "stop",
		Short: "Stop services",
		Long: `Stop running containers without removing them.

They can be started again with "start".`,
		Args: cobra.ArbitraryArgs,
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
	AddCommandBuilder(newStopCommand)
}
