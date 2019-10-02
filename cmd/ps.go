package cmd

import (
	"github.com/spf13/cobra"

	"gerrit.instructure.com/muss/config"
)

func newPsCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "ps",
		Short: "List containers",
		Args:  cobra.ArbitraryArgs,
		// TODO: ArgsInUseLine: "[service...]"
		RunE: func(cmd *cobra.Command, args []string) error {
			config.Save()
			return DelegateCmd(
				cmd,
				dockerComposeCmd(cmd, args),
			)
		},
	}

	cmd.Flags().BoolP("quiet", "q", false, "Only display IDs.")
	cmd.Flags().BoolP("services", "", false, "Display service names.")
	cmd.Flags().StringP("filter", "", "", "Filter services by a property.")
	cmd.Flags().BoolP("all", "a", false, "Show all stopped containers (including those created by the run command).")

	return cmd
}

func init() {
	rootCmd.AddCommand(newPsCommand())
}
