package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"regexp"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"gerrit.instructure.com/muss/config"
	"gerrit.instructure.com/muss/proc"
	"gerrit.instructure.com/muss/term"
)

var dc = "docker-compose"

func configSavePreRun(cfg *config.ProjectConfig) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, argv []string) error {
		err := cfg.Save()
		for _, w := range cfg.Warnings {
			fmt.Fprintln(cmd.ErrOrStderr(), w)
		}
		return QuietErrorOrNil(err)
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

func dockerComposeArgs(action string, cmd *cobra.Command, args []string) []string {
	flags := (flagDumper{}).fromCmd(cmd)

	cmdargs := make([]string, 1, 1+len(flags)+len(args))
	cmdargs[0] = action

	cmdargs = append(cmdargs, flags...)
	cmdargs = append(cmdargs, args...)

	return cmdargs
}

func dockerComposeCmd(cmd *cobra.Command, args []string) *exec.Cmd {
	return dockerComposeNamedCmd(cmd.CalledAs(), cmd, args)
}

func dockerComposeNamedCmd(action string, cmd *cobra.Command, args []string) *exec.Cmd {
	return exec.Command(dc, dockerComposeArgs(action, cmd, args)...)
}

func dockerComposeExec(cmd *cobra.Command, args []string) error {
	return proc.Exec(append([]string{dc}, dockerComposeArgs(cmd.CalledAs(), cmd, args)...))
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

type dcErrorFilter struct {
	*proc.Pipe
	cfg          *config.ProjectConfig
	messages     []string
	readerDoneCh chan bool
}

func newDCErrorFilter(cfg *config.ProjectConfig) proc.StreamFilter {
	return &dcErrorFilter{
		cfg:          cfg,
		readerDoneCh: make(chan bool, 1),
		Pipe:         &proc.Pipe{},
	}
}

var rePullingImage = regexp.MustCompile(`^Pulling (\S+) \((?:https?://)?([^/]+)`)
var reGetNoBasicAuth = regexp.MustCompile(`Get (?:https?://)?([^/]+)\S+?: no basic auth credentials`)
var reParsingHTTP403 = regexp.MustCompile(`(?:Service '(\S+)' failed to build:\s+|for (\S+)\s+)?error parsing HTTP 403 response body: unexpected end of JSON input: ""`)
var reNoStoredCredential = regexp.MustCompile(`No stored credential for ([^"]+)`)

func (f *dcErrorFilter) Start(doneCh chan bool) {
	cfg := f.cfg
	reader := f.Reader()
	writer := f.Writer()
	f.messages = make([]string, 0)

	// Collect registries that have login errors so that at the end we only print
	// once for each registry, and only print "unknown registry" if we couldn't
	// identify any specific ones.
	// The output for `dc pull` with a 403 will show the error once with
	// each service name and then print them all again without.
	unknownRegistryLogin := false
	registriesForLogin := make([]string, 0)

	var lastRegistry, lastService string

	go func() {
		if f, ok := reader.(io.ReadCloser); ok {
			defer f.Close()
		}
		scanner := bufio.NewScanner(reader)
		scanner.Split(term.ScanLinesOrAnsiMovements)
		for scanner.Scan() {
			line := scanner.Bytes()
			length := len(line)
			// Only do the regexp matching for full lines.
			if length > 0 && line[length-1] == '\n' {

				if match := rePullingImage.FindSubmatch(line); match != nil {
					lastService = string(match[1])
					lastRegistry = string(match[2])
				} else if match := reNoStoredCredential.FindSubmatch(line); match != nil {
					registriesForLogin = append(registriesForLogin, string(match[1]))
				} else if match := reGetNoBasicAuth.FindSubmatch(line); match != nil {
					registriesForLogin = append(registriesForLogin, string(match[1]))
				} else if match := reParsingHTTP403.FindSubmatch(line); match != nil {
					registry := lastRegistry
					if len(match[1]) > 0 {
						registry = registryFromImage(dcImageForService(cfg, string(match[1])))
					} else if len(match[2]) > 0 {
						registry = registryFromImage(dcImageForService(cfg, string(match[2])))
					} else if lastRegistry == "" && lastService != "" {
						registry = registryFromImage(dcImageForService(cfg, lastService))
					}
					if registry == "" {
						unknownRegistryLogin = true
					} else {
						registriesForLogin = append(registriesForLogin, registry)
					}
				} else {
					lastService = ""
					lastRegistry = ""
				}

			}
			// Propagate to STDERR.
			writer.Write(line)
		}

		loginMessageFormat := "You may need to login to %s"
		if len(registriesForLogin) > 0 {
			for _, registry := range registriesForLogin {
				f.messages = appendOnce(f.messages, fmt.Sprintf(loginMessageFormat, registry))
			}
		} else if unknownRegistryLogin {
			f.messages = appendOnce(f.messages, fmt.Sprintf(loginMessageFormat, "your docker registry"))
		}

		f.readerDoneCh <- true
	}()
}

func appendOnce(slice []string, add string) []string {
	for _, s := range slice {
		if s == add {
			return slice
		}
	}
	return append(slice, add)
}

func dcImageForService(cfg *config.ProjectConfig, svc string) string {
	if cfg == nil {
		return ""
	}
	dc, err := cfg.ComposeConfig()
	if err != nil {
		return ""
	}
	if image, ok := subMap(dc, "services", svc)["image"].(string); ok {
		return image
	}
	return ""
}

func subMap(m map[string]interface{}, keys ...string) map[string]interface{} {
	var ok bool
	for _, k := range keys {
		if m, ok = m[k].(map[string]interface{}); !ok {
			return map[string]interface{}{}
		}
	}
	return m
}

var reRegistryFromImage = regexp.MustCompile(`^(?:https?://)?([^/]+)/[^/]+/`)

func registryFromImage(image string) string {
	if match := reRegistryFromImage.FindStringSubmatch(image); match != nil {
		return match[1]
	}
	return ""
}

func (f *dcErrorFilter) Stop() {
	// Wait until the filter is complete.
	<-f.readerDoneCh

	if len(f.messages) == 0 {
		return
	}
	writer := f.Writer()
	// Print a spacer line to separate pass-through errors from our messages.
	fmt.Fprintln(writer, "")
	for _, msg := range f.messages {
		fmt.Fprintln(writer, msg)
	}
}
