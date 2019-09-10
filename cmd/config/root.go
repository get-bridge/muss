package config

import (
	"github.com/spf13/cobra"
)

// NewCommand builds the config subcommand.
func NewCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "config",
		Short: "muss configuration",
		Long:  `Work with muss configuration.`,
	}

	cmd.AddCommand(
		newSaveCommand(),
		newShowCommand(),
	)

	return cmd
}
