package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync"
)

// EnvCommand is a command that sets (an) env var(s).
type EnvCommand struct {
	Exec    []string `yaml:"exec"`
	Parse   bool     `yaml:"parse"`
	Varname string   `yaml:"varname"`
}

type envLoader interface {
	ShouldParse() bool
	Value() ([]byte, error)
	VarName() string
}

// LoadEnv will load environment variables from all config sources
// including project_name and secret commands.
func (cfg *ProjectConfig) LoadEnv() error {
	if cfg.ProjectName != "" {
		setenvIfUnset("COMPOSE_PROJECT_NAME", cfg.ProjectName)
	}

	if cfg.ComposeFile != "" {
		setenvIfUnset("COMPOSE_FILE", cfg.ComposeFile)
	}

	if err := loadEnvFromCmds(cfg.Secrets...); err != nil {
		return fmt.Errorf("Failed to load secrets: %w", err)
	}

	return nil
}

// ShouldParse is true if the output should be parsed and false if varname
// should be used.
func (e *EnvCommand) ShouldParse() bool {
	return e.Parse
}

// Value will run the command and return the output.
func (e *EnvCommand) Value() ([]byte, error) {
	var stdout bytes.Buffer
	cmd := exec.Command(e.Exec[0], e.Exec[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = &stdout
	// Pass stderr to show password prompts (or any problems).
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("command failed: %s", err)
	}

	return bytes.TrimRight(stdout.Bytes(), "\n"), nil
}

// VarName returns the name of the env var that the command will set.
func (e *EnvCommand) VarName() string {
	return e.Varname
}

func loadEnv(e envLoader) error {
	// For a single value...
	if !e.ShouldParse() {
		varname := e.VarName()
		if varname == "" {
			return errors.New(`env command must have either "parse: true" or a "varname"`)
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
			return errors.New(`use "parse: true" or "varname", not both`)
		}
		// If we don't know what env vars it will load
		// we have to call it.
		val, err := e.Value()
		if err != nil {
			return err
		}
		return loadEnvFromBytes(val)
	}
	return nil
}

// loadEnvFromCmds takes envLoaders and runs them and updates the current env.
func loadEnvFromCmds(envCmds ...envLoader) error {
	cmdErrors := make(chan error, len(envCmds))
	var wg sync.WaitGroup
	for _, env := range envCmds {
		wg.Add(1)
		go func(env envLoader) {
			defer wg.Done()
			err := loadEnv(env)
			if err != nil {
				cmdErrors <- err
			}
		}(env)
	}
	wg.Wait()
	close(cmdErrors)

	if len(cmdErrors) > 0 {
		var errorMessage string
		for e := range cmdErrors {
			errorMessage += e.Error()
		}
		return errors.New(errorMessage)
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

		setenvIfUnset(string(parts[0]), string(parts[1]))
	}

	return nil
}

func setenvIfUnset(key string, value string) (err error) {
	if _, ok := os.LookupEnv(key); !ok {
		err = os.Setenv(key, value)
	}
	return
}
