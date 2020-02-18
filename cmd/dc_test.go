package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gerrit.instructure.com/muss/proc"
)

func TestDcCommand(t *testing.T) {
	withTestPath(t, func(t *testing.T) {
		t.Run("all args pass through", func(t *testing.T) {
			_, _, err := runTestCommand(nil, []string{
				"dc",
				"--no-ansi",
				"down",
				"-v",
			})

			assert.Nil(t, err)

			assert.Equal(t,
				[]string{"docker-compose", "--no-ansi", "down", "-v"},
				proc.LastExecArgv,
				"exec")
		})
	})
}
