package config

import (
	"github.com/spf13/cobra"

	rootcmd "gerrit.instructure.com/muss/cmd"
	"gerrit.instructure.com/muss/config"
)

func newSaveCommand(cfg *config.ProjectConfig) *cobra.Command {
	target := cfg.ComposeFilePath()

	var saveCmd = &cobra.Command{
		Use:   "save",
		Short: "Generate new config files",
		Long:  `Generate new ` + target + ` file.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cfg.Save()
			return rootcmd.QuietErrorOrNil(err)
		},
	}
	return saveCmd
}

func init() {
	AddCommandBuilder(newSaveCommand)
}
