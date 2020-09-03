package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/instructure-bridge/muss/config"
	"github.com/instructure-bridge/muss/testutil"
)

func testRootCmd(args ...string) (int, string, string) {
	var stdout, stderr strings.Builder

	cfg, _ := config.NewConfigFromMap(nil)
	rootCmd := NewRootCommand(cfg)

	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)

	exitCode := ExecuteRoot(rootCmd, args)

	return exitCode, stdout.String(), stderr.String()
}

func getLines(s string, want int) []string {
	lines := strings.SplitAfter(s, "\n")
	end := len(lines)
	if want < end {
		end = want
	}
	return lines[0:end]
}

func TestRootCommand(t *testing.T) {
	withTestPath(t, func(t *testing.T) {
		t.Run("bad flag", func(t *testing.T) {
			exitCode, stdout, stderr := testRootCmd("--foo")

			assert.Equal(t, 1, exitCode, "exit 1")
			assert.Equal(t, "", stdout)
			assert.Equal(t,
				[]string{"Error:  unknown flag: --foo\n", "\n", "Usage:\n", "  muss [command]\n"},
				getLines(stderr, 4),
			)
		})

		t.Run("bad subcmd", func(t *testing.T) {
			exitCode, stdout, stderr := testRootCmd("foo")

			assert.Equal(t, 1, exitCode, "exit 1")
			assert.Equal(t, "", stdout)
			assert.Equal(t,
				[]string{"Error:  unknown command \"foo\" for \"muss\"\n", "\n", "Usage:\n", "  muss [command]\n"},
				getLines(stderr, 4),
			)
		})

		t.Run("bad subcmd flag", func(t *testing.T) {
			exitCode, stdout, stderr := testRootCmd("wrap", "--foo")

			assert.Equal(t, 1, exitCode, "exit 1")
			assert.Equal(t, "", stdout)
			assert.Equal(t,
				[]string{"Error:  unknown flag: --foo\n", "\n", "Usage:\n", "  muss wrap [flags]\n"},
				getLines(stderr, 4),
			)
		})

		t.Run("non-zero delegated command exit", func(t *testing.T) {
			os.Setenv("MUSS_TEST_DC_ERROR", "2")
			defer os.Unsetenv("MUSS_TEST_DC_ERROR")
			exitCode, stdout, stderr := testRootCmd("pull")

			assert.Equal(t, 2, exitCode, "exit 2")
			assert.Equal(t, "", stdout)
			assert.Equal(t, "", stderr)
		})

		t.Run("success", func(t *testing.T) {
			exitCode, stdout, stderr := testRootCmd("pull")

			assert.Equal(t, 0, exitCode, "exit 0")
			assert.Equal(t, "docker-compose\npull\n", stdout)
			assert.Equal(t, "std err\n", stderr)
		})
	})

	t.Run("Execute()", func(t *testing.T) {
		testutil.WithTempDir(t, func(tmpdir string) {
			yaml := `---
module_definitions:
- name: foo
  configs:
    sole:
      version: "1.5"
`
			testutil.WriteFile(t, "muss.yaml", yaml)
			dest := "docker-compose.yml"
			testutil.NoFileExists(t, dest)

			assert.Equal(t, 0, Execute([]string{"wrap", "true"}), "exit 0")

			assert.Contains(t, testutil.ReadFile(t, dest), `version: "1.5"`, "config written")
			os.Remove(dest)
			testutil.NoFileExists(t, dest)

			assert.Equal(t, 1, Execute([]string{"wrap", "false"}), "exit 1")
			assert.Contains(t, testutil.ReadFile(t, dest), `version: "1.5"`, "config written again")
		})
	})
}
