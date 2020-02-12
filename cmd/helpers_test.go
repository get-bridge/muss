package cmd

import (
	"os"
	"path"
	"strings"
	"testing"

	"gerrit.instructure.com/muss/config"
)

// helpers

var testbin string

func init() {
	cwd, err := os.Getwd()
	if err != nil {
		panic("Failed to get current dir: " + err.Error())
	}
	testbin = path.Join(cwd, "..", "testdata", "bin")
}

func runTestCommand(cfg *config.ProjectConfig, args []string) (string, string, error) {
	var stdout, stderr strings.Builder

	cmd := NewRootCommand(cfg)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	config.SetStderr(&stderr)
	cmd.SetArgs(args)

	// Don't write config files.
	sub, _, _ := cmd.Find(args)
	sub.PreRun = nil

	err := cmd.Execute()

	return stdout.String(), stderr.String(), err
}

func withTestPath(t *testing.T, f func(*testing.T)) {
	path := os.Getenv("PATH")
	os.Setenv("PATH", strings.Join([]string{testbin, path}, string(os.PathListSeparator)))
	defer os.Setenv("PATH", path)
	t.Run("with test path", f)
}
