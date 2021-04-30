package cmd

import (
	"github.com/spf13/cobra"

	"github.com/get-bridge/muss/config"
)

func newRunCommand(cfg *config.ProjectConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "run",
		Short: "Run a one-off command",
		Long: `Run a one-off command on a service.

Use this to execute a new process inside a new container.

NOTE: muss run implies "--rm".
Pass "--no-rm" if you want the container to persist.

By default, linked services will be started, unless they are already
running. If you do not want to start linked services, use "--no-deps".

Usage: run [options] SERVICE [COMMAND] [ARGS...]

Options:
  -d, --detach          Detached mode: Run container in the background, print
                        new container name.
  --name NAME           Assign a name to the container
  --entrypoint CMD      Override the entrypoint of the image.
  -e KEY=VAL            Set an environment variable (can be used multiple times)
  -l, --label KEY=VAL   Add or override a label (can be used multiple times)
  -u, --user=""         Run as specified username or uid
  --no-deps             Don't start linked services.
  --no-rm               Don't remove container after run.
  -p, --publish=[]      Publish a container's port(s) to the host
  --service-ports       Run command with the service's ports enabled and mapped
                        to the host.
  --use-aliases         Use the service's network aliases in the network(s) the
                        container connects to.
  -v, --volume=[]       Bind mount a volume (default [])
  -T                    Disable pseudo-tty allocation. By default "run"
                        allocates a TTY.
  -w, --workdir=""      Working directory inside the container
`,
		Args:               cobra.ArbitraryArgs,
		DisableFlagParsing: true,
		PreRunE:            configSavePreRun(cfg),
		RunE: func(cmd *cobra.Command, args []string) error {

			// Pass --rm by default (allow --no-rm to disable).
			rm := true
			cmdargs := make([]string, 0, len(args)+1)
			for i, arg := range args {
				if arg == "--rm" {
					rm = true
				} else if arg == "--no-rm" {
					rm = false
				} else if arg == "--" {
					cmdargs = append(cmdargs, args[i:]...)
					break
				} else {
					cmdargs = append(cmdargs, arg)
				}
			}
			if rm {
				cmdargs = append([]string{"--rm"}, cmdargs...)
			}

			return dockerComposeExec(cmd, cmdargs)
		},
	}

	// Just show "Long" by providing a zero-width (but not zero-value) string.
	cmd.SetUsageTemplate(`{{ "" }}`)

	return cmd
}

func init() {
	AddCommandBuilder(newRunCommand)
}
