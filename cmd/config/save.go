package config

import (
	"github.com/spf13/cobra"

	config "gerrit.instructure.com/muss/config"
)

func newSaveCommand() *cobra.Command {
	var saveCmd = &cobra.Command{
		Use:   "save",
		Short: "Generate new config files",
		// TODO: Eventually this may include secrets files.
		Long: `Generate new ` + config.DockerComposeFile + ` file.`,
		Run: func(cmd *cobra.Command, args []string) {
			config.Save()
		},
	}
	return saveCmd
}
