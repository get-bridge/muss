package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"gerrit.instructure.com/muss/config"
	"gerrit.instructure.com/muss/term"
)

func newUpCommand() *cobra.Command {
	opts := struct {
		noStatus bool

		detach               bool
		noColor              bool
		quietPull            bool
		noDeps               bool
		forceRecreate        bool
		alwaysRecreateDeps   bool
		noRecreate           bool
		noBuild              bool
		noStart              bool
		build                bool
		abortOnContainerExit bool
		timeout              int
		renewAnonVolumes     bool
		removeOrphans        bool
		exitCodeFrom         string
		scale                string
	}{}

	var cmd = &cobra.Command{
		Use:   "up",
		Short: "Create and start containers",
		Long: `Builds, (re)creates, starts, and attaches to containers for a service.

Unless they are already running, this command also starts any linked services.

The "up" command aggregates the output of each container. When
the command exits, all containers are stopped. Running "up -d"
starts the containers in the background and leaves them running.

If there are existing containers for a service, and the service's configuration
or image was changed after the container's creation, "up" picks
up the changes by stopping and recreating the containers (preserving mounted
volumes). To prevent picking up changes, use the "--no-recreate" flag.

If you want to force Compose to stop and recreate all containers, use the
"--force-recreate" flag.
`,
		Args:   cobra.ArbitraryArgs,
		PreRun: configSavePreRun,
		RunE: func(cmd *cobra.Command, args []string) error {

			switch {
			// TODO: global noANSI
			case opts.detach:
				fallthrough
			case opts.noStart:
				fallthrough
			case opts.noStatus:
				return DelegateCmd(
					cmd,
					dockerComposeCmd(cmd, args),
				)
			default:
			}

			return runUpWithStatus(cmd, args)
		},
	}

	cmd.Flags().SortFlags = false
	// muss only
	// TODO: need to annotate custom muss flags so that we don't pass them to
	// docker-compose.
	// cmd.Flags().BoolVarP(&opts.noStatus, "no-status", "", false, "Do not show muss status at the bottom of the log output.")

	cmd.Flags().BoolVarP(&opts.detach, "detach", "d", false, "Detached mode: Run containers in the background,\nprint new container names. Incompatible with\n--abort-on-container-exit.")
	cmd.Flags().BoolVarP(&opts.noColor, "no-color", "", false, "Produce monochrome output.")
	cmd.Flags().BoolVarP(&opts.quietPull, "quiet-pull", "", false, "Pull without printing progress information")
	cmd.Flags().BoolVarP(&opts.noDeps, "no-deps", "", false, "Don't start linked services.")
	cmd.Flags().BoolVarP(&opts.forceRecreate, "force-recreate", "", false, "Recreate containers even if their configuration\nand image haven't changed.")
	cmd.Flags().BoolVarP(&opts.alwaysRecreateDeps, "always-recreate-deps", "", false, "Recreate dependent containers.\nIncompatible with --no-recreate.")
	cmd.Flags().BoolVarP(&opts.noRecreate, "no-recreate", "", false, "If containers already exist, don't recreate\nthem. Incompatible with --force-recreate and -V.")
	cmd.Flags().BoolVarP(&opts.noBuild, "no-build", "", false, "Don't build an image, even if it's missing.")
	cmd.Flags().BoolVarP(&opts.noStart, "no-start", "", false, "Don't start the services after creating them.")
	cmd.Flags().BoolVarP(&opts.build, "build", "", false, "Build images before starting containers.")
	cmd.Flags().BoolVarP(&opts.abortOnContainerExit, "abort-on-container-exit", "", false, "Stops all containers if any container was\nstopped. Incompatible with -d.")
	cmd.Flags().IntVarP(&opts.timeout, "timeout", "t", 10, "Use this `timeout` in seconds for container\nshutdown when attached or when containers are\nalready running. (default: 10)")
	cmd.Flags().BoolVarP(&opts.renewAnonVolumes, "renew-anon-volumes", "V", false, "Recreate anonymous volumes instead of retrieving\ndata from the previous containers.")
	cmd.Flags().BoolVarP(&opts.removeOrphans, "remove-orphans", "", false, "Remove containers for services not defined\nin the Compose file.")
	cmd.Flags().StringVarP(&opts.exitCodeFrom, "exit-code-from", "", "", "Return the exit code of the selected `service`\ncontainer. Implies --abort-on-container-exit.")
	cmd.Flags().StringVarP(&opts.scale, "scale", "", "", "With `SERVICE=NUM` scale SERVICE to NUM instances.\nOverrides the `scale` setting in the Compose file if present.")

	return cmd
}

func init() {
	rootCmd.AddCommand(newUpCommand())
}

func runUpWithStatus(cmd *cobra.Command, args []string) error {
	cfg, err := config.All()
	if err != nil {
		return err
	}

	pr, pw, err := os.Pipe()
	if err != nil {
		return err
	}

	delegator := cmdDelegator(cmd)
	delegator.Stdout = pw

	done := make(chan bool)
	delegator.DoneCh = done

	// Setup a channel for log output.
	outputCh := make(chan []byte, 10)
	go func() {
		defer pr.Close()
		scanner := bufio.NewScanner(pr)
		for scanner.Scan() {
			outputCh <- scanner.Bytes()
		}
	}()

	// Setup a channel for status updates.
	statusCh := make(chan string, 1)
	statusCh <- "# muss"

	if cfg != nil && cfg.Status != nil && len(cfg.Status.Exec) > 0 {
		go func() {
			format := "# %s"
			if cfg.Status.LineFormat != "" {
				format = cfg.Status.LineFormat
			}
			statusCmd := cfg.Status.Exec
			interval := cfg.Status.Interval

			var stdout bytes.Buffer
			for {
				select {
				// After() should be fine here since there are only two paths:
				// either the timer has finished or we are exiting.
				case <-time.After(interval):
					stdout.Reset()
					cmd := exec.Command(statusCmd[0], statusCmd[1:]...)
					cmd.Stdin = os.Stdin
					cmd.Stdout = &stdout
					cmd.Run()
					lines := strings.Split(strings.TrimRight(stdout.String(), "\n"), "\n")
					for i, line := range lines {
						lines[i] = fmt.Sprintf(format, line)
					}
					statusCh <- strings.Join(lines, "\n")
				case <-done:
					return // go routine
				}
			}
		}()
	}

	go term.WriteWithFixedStatusLine(cmd.OutOrStdout(), outputCh, statusCh, done)

	defer func() {
		pw.Close()
	}()

	return delegator.Delegate(
		dockerComposeCmd(cmd, args),
	)
}
