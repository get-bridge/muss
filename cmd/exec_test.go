package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/instructure-bridge/muss/proc"
)

func TestExecCommand(t *testing.T) {
	withTestPath(t, func(t *testing.T) {
		t.Run("all args pass through", func(t *testing.T) {
			stdout, stderr, err := runTestCommand(nil, []string{
				"exec",
				"-d",
				"--privileged",
				"-u", "root",
				"-T",
				"--index=2",
				"-e", "k=some val",
				"--env", "k2=val2",
				"-w", "/dir",
				"svc",
				"cmd",
				"arg 1",
				"arg 2",
			})

			exp := []string{
				"docker-compose",
				"exec",
				"-d",
				"--privileged",
				"-u",
				"root",
				"-T",
				"--index=2",
				"-e",
				"k=some val",
				"--env",
				"k2=val2",
				"-w",
				"/dir",
				"svc",
				"cmd",
				"arg 1",
				"arg 2",
			}

			assert.Nil(t, err)
			assert.Equal(t, "", stderr, "exec not actually called")
			assert.Equal(t, "", stdout, "exec not actually called")

			assert.Equal(t, exp, proc.LastExecArgv)
		})
	})
}
