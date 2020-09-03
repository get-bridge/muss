package cmd

import (
	"os"
	"path"
	"strings"
	"testing"

	"github.com/instructure-bridge/muss/config"
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

func newTestConfig(t *testing.T, cfgMap map[string]interface{}) *config.ProjectConfig {
	cfg, err := config.NewConfigFromMap(cfgMap)
	if err != nil {
		t.Fatalf("unexpected config error: %s", err)
	}
	return cfg
}

func runTestCommand(cfg *config.ProjectConfig, args []string) (string, string, error) {
	var stdout, stderr strings.Builder

	if cfg == nil {
		cfg, _ = config.NewConfigFromMap(nil)
	}

	cmd := NewRootCommand(cfg)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(args)

	// Don't write config files.
	sub, _, _ := cmd.Find(args)
	sub.PreRunE = nil

	err := cmd.Execute()

	return stdout.String(), stderr.String(), err
}

func withTestPath(t *testing.T, f func(*testing.T)) {
	path := os.Getenv("PATH")
	os.Setenv("PATH", strings.Join([]string{testbin, path}, string(os.PathListSeparator)))
	defer os.Setenv("PATH", path)
	t.Run("with test path", f)
}
