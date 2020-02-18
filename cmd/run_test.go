package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gerrit.instructure.com/muss/proc"
)

func TestRunCommand(t *testing.T) {
	withTestPath(t, func(t *testing.T) {
		t.Run("all args pass through", func(t *testing.T) {
			args := testRunArgs(t, []string{
				"-d",
				"--name", "foo",
				"--entrypoint=/bin/sh",
				"-e", "k=some value",
				"-u", "root",
				"--no-deps",
				"-v", "/here:/there",
				"-v", "another:/vol",
				"-T",
				"--workdir", "/opt",
				"svc",
				"echo",
				"foo",
			})

			exp := []string{
				"docker-compose",
				"run",
				"--rm",
				"-d",
				"--name",
				"foo",
				"--entrypoint=/bin/sh",
				"-e",
				"k=some value",
				"-u",
				"root",
				"--no-deps",
				"-v",
				"/here:/there",
				"-v",
				"another:/vol",
				"-T",
				"--workdir",
				"/opt",
				"svc",
				"echo",
				"foo",
			}

			assert.Equal(t, exp, args)
		})

		t.Run("--no-rm", func(t *testing.T) {
			args := testRunArgs(t, []string{
				"-d",
				"--no-rm",
				"svc",
				"echo",
				"foo",
			})

			exp := []string{
				"docker-compose",
				"run",
				"-d",
				"svc",
				"echo",
				"foo",
			}

			assert.Equal(t, exp, args)
		})

		t.Run("--no-rm after --", func(t *testing.T) {
			args := testRunArgs(t, []string{
				"-d",
				"--",
				"svc",
				"echo",
				"--no-rm",
				"foo",
			})

			exp := []string{
				"docker-compose",
				"run",
				"--rm",
				"-d",
				"--",
				"svc",
				"echo",
				"--no-rm",
				"foo",
			}

			assert.Equal(t, exp, args)
		})
	})
}

func testRunArgs(t *testing.T, args []string) []string {
	stdout, stderr, err := runTestCommand(nil, append([]string{"run"}, args...))

	assert.Nil(t, err)
	assert.Equal(t, "", stderr, "exec not actually get called")
	assert.Equal(t, "", stdout, "exec not actually get called")

	return proc.LastExecArgv
}
