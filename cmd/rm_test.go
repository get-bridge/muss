package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRmCommand(t *testing.T) {
	withTestPath(t, func(t *testing.T) {
		// Test that any valid dc args pass through.
		t.Run("all args pass through", func(t *testing.T) {
			stdout, stderr, err := runTestCommand(nil, []string{
				"rm",
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

		t.Run("no args", func(t *testing.T) {
			stdout, stderr, err := runTestCommand(nil, []string{"rm"})

			expOut := `docker-compose
rm
`

			assert.Nil(t, err)
			assert.Equal(t, "std err\n", stderr)
			assert.Equal(t, expOut, stdout)
		})
	})
}
