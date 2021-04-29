package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/get-bridge/muss/config"
)

func newAttachCommand(cfg *config.ProjectConfig) *cobra.Command {
	index := 1
	var cmd = &cobra.Command{
		Use:   "attach",
		Short: "Attach local stdio to a running container",
		// Long
		Args: cobra.ExactArgs(1),
		// TODO: ArgsInUseLine: "service"
		PreRunE: configSavePreRun(cfg),
		RunE: func(cmd *cobra.Command, args []string) error {
			service := args[0]
			cid, err := dockerContainerID(service)
			if err != nil {
				return err
			}

			lines := strings.Split(cid, "\n")
			if len(lines) >= index {
				cid = lines[index-1]
			} else {
				return fmt.Errorf("Index %d not found for service %s", index, service)
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

	cmd.Flags().SortFlags = false

	cmd.Flags().IntVarP(&index, "index", "", 1, "index of the container if there are multiple\ninstances of a service")
	cmd.Flags().SetAnnotation("index", "muss-only", []string{"true"})

	cmd.Flags().StringP("detach-keys", "", "ctrl-c", "Override the key sequence for detaching a container")
	cmd.Flags().BoolP("no-stdin", "", false, "Do not attach STDIN")
	cmd.Flags().BoolP("sig-proxy", "", false, "Proxy all received signals to the process")

	return cmd
}

func init() {
	AddCommandBuilder(newAttachCommand)
}
