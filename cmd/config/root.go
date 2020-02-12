package config

import (
	"fmt"

	"github.com/spf13/cobra"

	rootcmd "gerrit.instructure.com/muss/cmd"
	"gerrit.instructure.com/muss/config"
)

// CommandBuilder is a function that takes the project config as an argument
// and returns a cobra command.
type CommandBuilder func(*config.ProjectConfig) *cobra.Command

var cmdBuilders = make([]CommandBuilder, 0)

// AddCommandBuilder takes the provided function and adds it to the list of
// commands that will be added to the root command when it is built.
func AddCommandBuilder(f CommandBuilder) {
	cmdBuilders = append(cmdBuilders, f)
}

// NewCommand builds the config subcommand.
func NewCommand(cfg *config.ProjectConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "config",
		Short: "muss configuration commands",
		Long:  `Work with muss configuration.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if cfg, err := config.All(); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "error loading config: %s", err)
			} else {
				if cfg == nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "muss project config '%s' file not found.\n", config.ProjectFile)
				}
			}
		},
	}

	for _, f := range cmdBuilders {
		cmd.AddCommand(f(cfg))
	}

	return cmd
}

func init() {
	rootcmd.AddCommandBuilder(NewCommand)
}
