package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPsCommand(t *testing.T) {
	withTestPath(t, func(t *testing.T) {
		// Test that any valid dc args pass through.
		t.Run("all args pass through", func(t *testing.T) {
			stdout, stderr, err := runTestCommand(nil, []string{
				"ps",
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

		t.Run("some args", func(t *testing.T) {
			stdout, stderr, err := runTestCommand(nil, []string{
				"ps",
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
