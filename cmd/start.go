package cmd

import (
	"github.com/spf13/cobra"

	"gerrit.instructure.com/muss/config"
)

func newStartCommand(cfg *config.ProjectConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "start",
		Short: "Start services",
		Long:  "Start existing containers.",
		Args:  cobra.ArbitraryArgs,
		// TODO: ArgsInUseLine: "[service...]"
		PreRunE: configSavePreRun(cfg),
		RunE: func(cmd *cobra.Command, args []string) error {
			return DelegateCmd(
				cmd,
				dockerComposeCmd(cmd, args),
			)
		},
	}

	return cmd
}

func init() {
	AddCommandBuilder(newStartCommand)
}
