package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRmCommand(t *testing.T) {
	withTestPath(t, func(*testing.T) {
		// Test that any valid dc args pass through.
		t.Run("all args pass through", func(*testing.T) {
			stdout, stderr, err := testCmdBuilder(newRmCommand, []string{
				"-f",
				"-s",
				"-v",
				"svc",
			})

			// sorted and normalized
			expOut := `docker-compose
rm
--force
--stop
-v
svc
`

			assert.Nil(t, err)
			assert.Equal(t, "std err\n", stderr)
			assert.Equal(t, expOut, stdout)
		})

		t.Run("no args", func(*testing.T) {
			stdout, stderr, err := testCmdBuilder(newRmCommand, []string{})

			expOut := `docker-compose
rm
`

			assert.Nil(t, err)
			assert.Equal(t, "std err\n", stderr)
			assert.Equal(t, expOut, stdout)
		})
	})
}
