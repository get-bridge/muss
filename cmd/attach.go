package cmd

import (
	"github.com/spf13/cobra"
)

func newAttachCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "attach",
		Short: "Attach local stdio to a running container",
		// Long
		Args: cobra.ExactArgs(1),
		// TODO: ArgsInUseLine: "service"
		PreRun: configSavePreRun,
		RunE: func(cmd *cobra.Command, args []string) error {
			cid, err := dockerContainerID(args[0])
			if err != nil {
				return err
			}

			cmdArgs := []string{"attach"}
			cmdArgs = append(cmdArgs, (flagDumper{visitAll: true, showFalseBools: true}).fromCmd(cmd)...)
			cmdArgs = append(cmdArgs, cid)

			return DelegateCmd(
				cmd,
				dockerCmd(cmdArgs...),
			)
		},
	}

	cmd.Flags().StringP("detach-keys", "", "ctrl-c", "Override the key sequence for detaching a container")
	cmd.Flags().BoolP("no-stdin", "", false, "Do not attach STDIN")
	cmd.Flags().BoolP("sig-proxy", "", false, "Proxy all received signals to the process")

	return cmd
}

func init() {
	rootCmd.AddCommand(newAttachCommand())
}
