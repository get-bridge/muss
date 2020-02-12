package cmd

import (
	"github.com/spf13/cobra"

	"gerrit.instructure.com/muss/config"
)

func newLogsCommand(cfg *config.ProjectConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "logs",
		Short: "View output from services",
		// Long
		Args: cobra.ArbitraryArgs,
		// TODO: ArgsInUseLine: "[service...]"
		PreRunE: configSavePreRun(cfg),
		RunE: func(cmd *cobra.Command, args []string) error {
			return DelegateCmd(
				cmd,
				dockerComposeCmd(cmd, args),
			)
		},
	}

	cmd.Flags().BoolP("no-color", "", false, "Produce monochrome output.")
	cmd.Flags().BoolP("follow", "f", false, "Follow log output.")
	cmd.Flags().BoolP("timestamps", "t", false, "Show timestamps.")
	cmd.Flags().StringP("tail", "", "all", "Number of lines to show from the end of the logs for each container.")

	return cmd
}

func init() {
	AddCommandBuilder(newLogsCommand)
}
