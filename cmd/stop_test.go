package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStopCommand(t *testing.T) {
	withTestPath(t, func(*testing.T) {
		// Test that any valid dc args pass through.
		t.Run("all args pass through", func(*testing.T) {
			stdout, stderr, err := testCmdBuilder(newStopCommand, []string{
				"-t", "11",
				"svc",
			})

			// sorted and normalized
			expOut := `docker-compose
stop
--timeout=11
svc
`

			assert.Nil(t, err)
			assert.Equal(t, "std err\n", stderr)
			assert.Equal(t, expOut, stdout)
		})

		t.Run("no args", func(*testing.T) {
			stdout, stderr, err := testCmdBuilder(newStopCommand, []string{})

			expOut := `docker-compose
stop
`

			assert.Nil(t, err)
			assert.Equal(t, "std err\n", stderr)
			assert.Equal(t, expOut, stdout)
		})
	})
}
