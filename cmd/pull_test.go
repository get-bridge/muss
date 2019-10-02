package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPullCommand(t *testing.T) {
	withTestPath(t, func(*testing.T) {
		t.Run("all args pass through", func(*testing.T) {
			stdout, stderr, err := testCmdBuilder(newPullCommand, []string{
				"--ignore-pull-failures",
				"--no-parallel",
				"-q",
				"--include-deps",
				"svc",
			})

			expOut := `docker-compose
pull
--ignore-pull-failures
--include-deps
--no-parallel
--quiet
svc
`

			assert.Nil(t, err)
			assert.Equal(t, "std err\n", stderr)
			assert.Equal(t, expOut, stdout)
		})

	})
}
