package cmd

import (
	"github.com/spf13/cobra"

	"gerrit.instructure.com/muss/config"
)

func newRmCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "rm",
		Short: "Remove stopped containers",
		Long: `Removes stopped service containers.

By default, anonymous volumes attached to containers will not be removed. You
can override this with "-v". To list all volumes, use "docker volume ls".

Any data which is not in a volume will be lost.`,
		Args: cobra.ArbitraryArgs,
		// TODO: ArgsInUseLine: "[service...]"
		RunE: func(cmd *cobra.Command, args []string) error {
			config.Save()
			return DelegateCmd(
				cmd,
				dockerComposeCmd(cmd, args),
			)
		},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().BoolP("force", "f", false, "Don't ask to confirm removal")
	cmd.Flags().BoolP("stop", "s", false, "Stop the containers, if required, before removing")
	// dc does not define a long name for -v.
	cmd.Flags().BoolP("v", "v", false, "Remove any anonymous volumes attached to containers")

	return cmd
}

func init() {
	rootCmd.AddCommand(newRmCommand())
}
