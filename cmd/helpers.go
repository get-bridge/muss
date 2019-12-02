package cmd

import (
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"gerrit.instructure.com/muss/proc"
)

var dc = "docker-compose"

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

func dcFlagsFromCmd(cmd *cobra.Command) []string {
	args := make([]string, 0)

	// Determine which flags were set and pass them on.
	cmd.Flags().Visit(func(flag *pflag.Flag) {
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
			// just the name
		case "int", "string":
			arg += "=" + flag.Value.String()
		default:
			panic("flag delegation undefined for " + flag.Name)
		}

		args = append(args, arg)
	})

	return args
}

func dockerComposeArgs(cmd *cobra.Command, args []string) []string {
	flags := dcFlagsFromCmd(cmd)

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
