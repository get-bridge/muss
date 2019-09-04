package cmd

import (
	"fmt"
	"github.com/spf13/cobra"

	cmdconfig "gerrit.instructure.com/muss/cmd/config"
	config "gerrit.instructure.com/muss/config"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "help",
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

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is "+config.DefaultFile+")")

	rootCmd.AddCommand(cmdconfig.NewCommand())
}

func initConfig() {
	if cfgFile != "" {
		config.ProjectFile = cfgFile
	}
}
