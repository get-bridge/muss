package config

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

type envCmd struct {
	parse   bool
	exec    []string
	varname string
}

type envLoader interface {
	ShouldParse() bool
	Value() ([]byte, error)
	VarName() string
}

// parseEnvCommands takes an item (interface{}) from a config
// (map[string]interface{}) and parses it into envCmd object(s) if possible.
func parseEnvCommands(spec interface{}) ([]envLoader, error) {
	commands := make([]envLoader, 0)
	if spec != nil {
		specs := make([]interface{}, 0)
		if slice, ok := spec.([]interface{}); ok {
			specs = append(specs, slice...)
		} else {
			specs = append(specs, spec)
		}
		for _, cmdSpec := range specs {
			cmd, err := parseEnvCommand(cmdSpec)
			if err != nil {
				return nil, err
			}
			commands = append(commands, cmd)
		}
	}

	return commands, nil
}

// parseEnvCommand attempts to parse a single interface{} into an envCmd.
func parseEnvCommand(spec interface{}) (*envCmd, error) {
	msi, ok := spec.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("expected env_command to be a map not %t", spec)
	}

	parse := false
	varname := ""
	args := []string{}
	for k, v := range msi {
		switch k {
		case "varname":
			varname = v.(string)
		case "parse":
			parse = v.(bool)
		case "exec":
			var ok bool
			args, ok = stringSlice(v)
			if !ok {
				return nil, fmt.Errorf("exec must be a list")
			}
		default:
			return nil, fmt.Errorf("unknown key for env_command: %s", k)
		}
	}

	return &envCmd{
		exec:    args,
		parse:   parse,
		varname: varname,
	}, nil
}

func (e *envCmd) ShouldParse() bool {
	return e.parse
}

// Value will run the command and return the output.
func (e *envCmd) Value() ([]byte, error) {
	var stdout bytes.Buffer
	cmd := exec.Command(e.exec[0], e.exec[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = &stdout
	// Pass stderr for to show password prompts (or any problems).
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("command failed: %s", err)
	}

	return bytes.TrimRight(stdout.Bytes(), "\n"), nil
}

func (e *envCmd) VarName() string {
	return e.varname
}

// loadEnvFromCmds takes envLoaders and runs them and updates the current env.
func loadEnvFromCmds(envCmds ...envLoader) error {
	// TODO: load these concurrently
	for _, e := range envCmds {
		// For a single value...
		if !e.ShouldParse() {
			varname := e.VarName()
			if varname == "" {
				return fmt.Errorf(`env command must have either "parse: true" or a "varname"`)
			}
			// Only get it if not already set.
			if _, ok := os.LookupEnv(varname); !ok {
				val, err := e.Value()
				if err != nil {
					return err
				}
				if err := os.Setenv(varname, string(val)); err != nil {
					return err
				}
			}
		} else {
			if e.VarName() != "" {
				return fmt.Errorf(`use "parse: true" or "varname", not both`)
			}
			// If we don't know what env vars it will load
			// we have to call it.
			val, err := e.Value()
			if err != nil {
				return err
			}
			return loadEnvFromBytes(val)
		}
	}

	return nil
}

func loadEnvFromBytes(env []byte) error {
	lines := bytes.Split(env, []byte("\n"))

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}

		parts := bytes.SplitN(line, []byte("="), 2)
		if len(parts) != 2 {
			return fmt.Errorf("failed to parse name=value line: %s", line)
		}

		name := string(parts[0])
		if _, ok := os.LookupEnv(name); !ok {
			err := os.Setenv(name, string(parts[1]))
			if err != nil {
				return err
			}
		}
	}

	return nil
}
