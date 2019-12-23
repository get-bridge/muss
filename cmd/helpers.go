package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"gerrit.instructure.com/muss/config"
	"gerrit.instructure.com/muss/proc"
)

var dc = "docker-compose"

func configSavePreRun(cmd *cobra.Command, argv []string) {
	if err := config.Save(); err != nil {
		fmt.Print(err)
		// Abort command (without printing usage).
		os.Exit(1)
	}
}

func cmdDelegator(cmd *cobra.Command) *proc.Delegator {
	return (&proc.Delegator{
		Stdin:  cmd.InOrStdin(),
		Stdout: cmd.OutOrStdout(),
		Stderr: cmd.ErrOrStderr(),
	})
}

// DelegateCmd runs with a delegator made from a `cobra.Cmd`.
func DelegateCmd(cmd *cobra.Command, commands ...*exec.Cmd) (err error) {
	return cmdDelegator(cmd).Delegate(commands...)
}

type flagDumper struct {
	visitAll       bool
	showFalseBools bool
}

func (f flagDumper) fromCmd(cmd *cobra.Command) []string {
	args := make([]string, 0)

	// Determine which flags were set and pass them on.
	flagToString := func(flag *pflag.Flag) {
		if flag.Name == "help" {
			return
		}
		if flag.Annotations != nil {
			if mussOnly := flag.Annotations["muss-only"]; len(mussOnly) == 1 && mussOnly[0] == "true" {
				return
			}
		}

		var arg string
		// If dc only defines the shorthand make sure we send it that way.
		// see also https://github.com/spf13/pflag/issues/213
		if flag.Name == flag.Shorthand {
			arg = "-" + flag.Shorthand
		} else {
			arg = "--" + flag.Name
		}

		switch flag.Value.Type() {
		case "bool":
			val := flag.Value.String()
			if val == "false" {
				if !f.showFalseBools {
					return
				}
				arg += "=" + val
			}
			// just the name
		case "int", "string":
			arg += "=" + flag.Value.String()
		default:
			panic("flag delegation undefined for " + flag.Name)
		}

		args = append(args, arg)
	}

	if f.visitAll {
		cmd.Flags().VisitAll(flagToString)
	} else {
		cmd.Flags().Visit(flagToString)
	}

	return args
}

func dockerCmd(args ...string) *exec.Cmd {
	return exec.Command("docker", args...)
}

func dockerComposeArgs(cmd *cobra.Command, args []string) []string {
	flags := (flagDumper{}).fromCmd(cmd)

	cmdargs := make([]string, 1, 1+len(flags)+len(args))
	cmdargs[0] = cmd.CalledAs()
	cmdargs = append(cmdargs, flags...)
	cmdargs = append(cmdargs, args...)

	return cmdargs
}

func dockerComposeCmd(cmd *cobra.Command, args []string) *exec.Cmd {
	return exec.Command(dc, dockerComposeArgs(cmd, args)...)
}

func dockerComposeExec(cmd *cobra.Command, args []string) error {
	return proc.Exec(append([]string{dc}, dockerComposeArgs(cmd, args)...))
}

func dockerContainerID(service string) (string, error) {
	cid, _, err := proc.CmdOutput("docker-compose", "ps", "-q", service)

	errorMessage := fmt.Sprintf("failed to get container id for %s", service)

	if err != nil {
		return "", fmt.Errorf("%s: %w", errorMessage, err)
	}

	if cid == "" {
		return "", fmt.Errorf("%s", errorMessage)
	}

	return cid, nil
}
