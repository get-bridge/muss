package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPsCommand(t *testing.T) {
	withTestPath(t, func(*testing.T) {
		// Test that any valid dc args pass through.
		t.Run("all args pass through", func(*testing.T) {
			stdout, stderr, err := testCmdBuilder(newPsCommand, []string{
				"-q",
				"--services",
				"--filter", "KEY=VAL",
				"-a",
				"svc",
			})

			// sorted and normalized
			expOut := `docker-compose
ps
--all
--filter=KEY=VAL
--quiet
--services
svc
`

			assert.Nil(t, err)
			assert.Equal(t, "std err\n", stderr)
			assert.Equal(t, expOut, stdout)
		})

		t.Run("some args", func(*testing.T) {
			stdout, stderr, err := testCmdBuilder(newPsCommand, []string{
				"--all",
			})

			expOut := `docker-compose
ps
--all
`

			assert.Nil(t, err)
			assert.Equal(t, "std err\n", stderr)
			assert.Equal(t, expOut, stdout)
		})
	})
}
