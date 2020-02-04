package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	cmdconfig "gerrit.instructure.com/muss/cmd/config"
	config "gerrit.instructure.com/muss/config"
)

var rootCmd = &cobra.Command{
	Use:   "muss",
	Short: "Configure and run project services",
	// Root command just shows help (which shows subcommands).
	// SilenceUsage and Errors so that we don't print excessively when dc exits non-zero.
	SilenceErrors: true,
	SilenceUsage:  true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(args []string) int {
	rootCmd.SetArgs(args)
	if err := rootCmd.Execute(); err != nil {
		// Propagate errors from command delegation.
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}

		stderr := rootCmd.ErrOrStderr()

		// Print information about other errors.
		fmt.Fprintln(stderr, "Error: ", err.Error())

		// Print usage if it's a flag error
		// TODO: Could this be any other type of error that we don't want to print usage for?
		cmd, _, findErr := rootCmd.Find(args)
		// If subcmd isn't found, print root command usage
		if findErr != nil {
			cmd = rootCmd
		}
		fmt.Fprintln(stderr, "") // Print blank line between "Error:" and "Usage:".
		fmt.Fprintln(stderr, cmd.UsageString())

		return 1
	}

	return 0
}

func init() {
	if cfgFile, ok := os.LookupEnv("MUSS_FILE"); ok {
		config.ProjectFile = cfgFile
	}
	config.UserFile = os.Getenv("MUSS_USER_FILE")

	rootCmd.AddCommand(cmdconfig.NewCommand())
}
