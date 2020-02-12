package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDownCommand(t *testing.T) {
	withTestPath(t, func(*testing.T) {
		t.Run("all args pass through", func(*testing.T) {
			stdout, stderr, err := runTestCommand(nil, []string{
				"down",
				"--rmi=local",
				"-v",
				"--remove-orphans",
				"-t", "10",
			})

			expOut := `docker-compose
down
--rmi=local
-v
--remove-orphans
-t
10
`

			assert.Nil(t, err)
			assert.Equal(t, "std err\n", stderr)
			assert.Equal(t, expOut, stdout)
		})
	})
}
