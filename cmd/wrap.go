package cmd

import (
	"errors"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/instructure-bridge/muss/config"
	"github.com/instructure-bridge/muss/proc"
)

func newWrapCommand(cfg *config.ProjectConfig) *cobra.Command {
	var shellCommands []string
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}
	useExec := false

	var cmd = &cobra.Command{
		Use:   "wrap",
		Short: "Execute arbitrary commands",
		Long: `Execute arbitrary commands after initializing muss.

Useful for testing project configuration, environment, and command execution.

Usage: wrap [options] [COMMAND ARGS...]`,
		Example: "  muss wrap bin/script args...",
		Args:    cobra.ArbitraryArgs,
		PreRunE: configSavePreRun(cfg),
		RunE: func(cmd *cobra.Command, args []string) error {

			if useExec {
				if len(shellCommands) > 0 {
					return errors.New("--exec and -c are mutually exclusive")
				}
				if len(args) < 1 {
					return errors.New("--exec requires a command")
				}

				return proc.Exec(args)
			}

			commands := make([]*exec.Cmd, 0, 1+len(shellCommands))
			if len(args) > 0 {
				commands = append(commands, exec.Command(args[0], args[1:]...))
			}
			for _, c := range shellCommands {
				commands = append(commands, exec.Command(shell, "-c", c))
			}

			return DelegateCmd(cmd, commands...)
		},
	}

	// The first non-option arg starts a command and its args
	// so don't parse any flags after that.
	cmd.Flags().SetInterspersed(false)

	cmd.Flags().StringArrayVarP(&shellCommands, "command", "c", []string{},
		"Additional command (run by the shell).  Can be specified multiple times.")
	cmd.Flags().BoolVarP(&useExec, "exec", "", false,
		"Use exec instead of built-in command delegation (mutually exclusive with -c).")
	cmd.Flags().StringVarP(&shell, "shell", "s", shell,
		"Shell to run -c commands (instead of $SHELL).\n")

	return cmd
}

func init() {
	AddCommandBuilder(newWrapCommand)
}
