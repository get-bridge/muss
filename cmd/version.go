package cmd

import (
	"fmt"
	"sync"

	"github.com/spf13/cobra"

	"github.com/get-bridge/muss/config"
	"github.com/get-bridge/muss/proc"
)

// Version is the program version, filled in from git during build process.
var Version = "[development]"

func newVersionCommand(_ *config.ProjectConfig) *cobra.Command {
	short := false
	var cmd = &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if short {
				fmt.Fprintln(cmd.OutOrStdout(), Version)
				return
			}

			// Work async since docker-compose is slow.
			var wg sync.WaitGroup

			getVersion := func(verVar *string, argv ...string) {
				wg.Add(1)
				go func() {
					stdout, stderr, err := proc.CmdOutput(argv...)
					if err != nil {
						*verVar = fmt.Sprintf("error getting %s version: %s\n%s", argv[0], err.Error(), stderr)
					} else {
						*verVar = stdout
					}
					wg.Done()
				}()
			}

			var dcVersion string
			getVersion(&dcVersion, "docker-compose", "version", "--short")

			var dockerVersions string
			getVersion(&dockerVersions, "docker", "version", "--format", `docker client {{ .Client.Version }}{{ "\n" }}docker server {{ .Server.Version }}`)

			wg.Wait()

			fmt.Fprintf(
				cmd.OutOrStdout(),
				"muss %s\ndocker-compose %s\n%s\n",
				Version,
				dcVersion,
				dockerVersions,
			)
		},
	}

	cmd.Flags().SortFlags = false
	cmd.Flags().BoolVarP(&short, "short", "", false, "Show only muss version number")

	return cmd
}

func init() {
	AddCommandBuilder(newVersionCommand)
}
