package cmd

import (
	"github.com/spf13/cobra"

	"gerrit.instructure.com/muss/config"
)

func newExecCommand(cfg *config.ProjectConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "exec",
		Short: "Execute a command in a running container",
		Long: `Execute a command in a running container.

Useful for interacting with the current state of a running container.

To use a container to run a new process, use the "run" command.

Usage: exec [options] SERVICE COMMAND [ARGS...]

Options:
  -d, --detach      Detached mode: Run command in the background.
  --privileged      Give extended privileges to the process.
  -u, --user USER   Run the command as this user.
  -T                Disable pseudo-tty allocation. By default "exec"
                    allocates a TTY.
  --index=index     index of the container if there are multiple
                    instances of a service [default: 1]
  -e, --env KEY=VAL Set environment variables (can be used multiple times,
                    not supported in API < 1.25)
  -w, --workdir DIR Path to workdir directory for this command.
`,
		Args:               cobra.ArbitraryArgs,
		DisableFlagParsing: true,
		PreRunE:            configSavePreRun(cfg),
		RunE: func(cmd *cobra.Command, args []string) error {

			return dockerComposeExec(cmd, args)
		},
	}

	// Just show "Long" by providing a zero-width (but not zero-value) string.
	cmd.SetUsageTemplate(`{{ "" }}`)

	return cmd
}

func init() {
	AddCommandBuilder(newExecCommand)
}
