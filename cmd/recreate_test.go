package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecreateCommand(t *testing.T) {
	withTestPath(t, func(t *testing.T) {
		// Test that any valid dc args pass through.
		t.Run("all args pass through", func(t *testing.T) {
			stdout, stderr, err := runTestCommand(nil, []string{
				"recreate",
				"--build",
				"--no-start",
				"--scale", "web=2",
				"-t", "11",
				"a", "b",
			})

			// sorted and normalized
			expOut := `docker-compose
up
--build
--no-start
--scale=web=2
--timeout=11
--detach
--force-recreate
--renew-anon-volumes
a
b
`

			assert.Nil(t, err)
			assert.Equal(t, "std err\n", stderr)
			assert.Equal(t, expOut, stdout)
		})

		t.Run("no args", func(t *testing.T) {
			stdout, stderr, err := runTestCommand(nil, []string{"recreate"})

			expOut := `docker-compose
up
--detach
--force-recreate
--renew-anon-volumes
`

			assert.Nil(t, err)
			assert.Equal(t, "std err\n", stderr)
			assert.Equal(t, expOut, stdout)
		})
	})
}
