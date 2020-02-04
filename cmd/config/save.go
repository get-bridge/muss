package config

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	config "gerrit.instructure.com/muss/config"
)

func newSaveCommand() *cobra.Command {
	target := "docker-compose.yml"
	cfg, _ := config.All()
	if cfg != nil {
		target = cfg.ComposeFilePath()
	}

	var saveCmd = &cobra.Command{
		Use:   "save",
		Short: "Generate new config files",
		Long:  `Generate new ` + target + ` file.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := config.Save(); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}
	return saveCmd
}
