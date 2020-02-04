package cmd

import (
	"os"
	"path"
	"strings"
	"testing"

	"github.com/spf13/cobra"

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

func testCmdBuilder(builder func() *cobra.Command, args []string) (string, string, error) {
	var stdout, stderr strings.Builder

	cmd := builder()
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	config.SetStderr(&stderr)
	cmd.SetArgs(args)

	// Don't write config files.
	cmd.PreRun = nil

	// The parent command sets these which this command would normally inherit.
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true

	err := cmd.Execute()

	return stdout.String(), stderr.String(), err
}

func withTestPath(t *testing.T, f func(*testing.T)) {
	path := os.Getenv("PATH")
	os.Setenv("PATH", strings.Join([]string{testbin, path}, string(os.PathListSeparator)))
	defer os.Setenv("PATH", path)
	t.Run("with test path", f)
}
