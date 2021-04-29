package cmd

import (
	"github.com/spf13/cobra"

	"github.com/get-bridge/muss/config"
)

func newPullCommand(cfg *config.ProjectConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "pull",
		Short: "Pull the latest images for services",
		Long: `Pulls docker images for defined services but does not start the containers.
`,
		Args: cobra.ArbitraryArgs,
		// TODO: ArgsInUseLine: "[service...]"
		PreRunE: configSavePreRun(cfg),
		RunE: func(cmd *cobra.Command, args []string) error {

			// TODO: pull repos

			delegator := cmdDelegator(cmd)
			err := delegator.FilterStderr(newDCErrorFilter(cfg))
			if err != nil {
				return err
			}
			return delegator.Delegate(
				dockerComposeCmd(cmd, args),
			)
		},
	}

	cmd.Flags().BoolP("ignore-pull-failures", "", false, "Pull what it can and ignores images with pull failures.")
	cmd.Flags().BoolP("no-parallel", "", false, "Disable parallel pulling.")
	cmd.Flags().BoolP("quiet", "q", false, "Pull without printing progress information.")
	cmd.Flags().BoolP("include-deps", "", false, "Also pull services declared as dependencies.")

	return cmd
}

func init() {
	AddCommandBuilder(newPullCommand)
}
