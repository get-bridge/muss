package config

import (
	"fmt"

	"github.com/spf13/cobra"

	config "gerrit.instructure.com/muss/config"
)

// NewCommand builds the config subcommand.
func NewCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "config",
		Short: "muss configuration",
		Long:  `Work with muss configuration.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if config.All() == nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "muss project config '%s' file not found.\n", config.ProjectFile)
			}
		},
	}

	cmd.AddCommand(
		newSaveCommand(),
		newShowCommand(),
	)

	return cmd
}
