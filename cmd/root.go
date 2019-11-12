package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cmdconfig "gerrit.instructure.com/muss/cmd/config"
	config "gerrit.instructure.com/muss/config"
)

var rootCmd = &cobra.Command{
	Use:   "muss",
	Short: "Configure and run project services",
	// Root command just shows help (which shows subcommands).
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() int {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		return 1
	}
	return 0
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.AddCommand(cmdconfig.NewCommand())
}

func initConfig() {
	if cfgFile, ok := os.LookupEnv("MUSS_FILE"); ok {
		config.ProjectFile = cfgFile
	}
	config.UserFile = os.Getenv("MUSS_USER_FILE")
}
