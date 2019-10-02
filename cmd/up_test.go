package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpCommand(t *testing.T) {
	withTestPath(t, func(*testing.T) {
		t.Run("all args pass through", func(*testing.T) {
			stdout, stderr, err := testCmdBuilder(newUpCommand, []string{
				"-d",
				"--no-color",
				"--quiet-pull",
				"--no-deps",
				"--force-recreate",
				"--always-recreate-deps",
				"--no-recreate",
				"--no-build",
				"--no-start",
				"--build",
				"--abort-on-container-exit",
				"-t", "4",
				"-V",
				"--remove-orphans",
				"--exit-code-from", "svc",
				"--scale", "SERVICE=NUM",
				"svc",
			})

			expOut := `docker-compose
up
-d
--no-color
--quiet-pull
--no-deps
--force-recreate
--always-recreate-deps
--no-recreate
--no-build
--no-start
--build
--abort-on-container-exit
-t
4
-V
--remove-orphans
--exit-code-from
svc
--scale
SERVICE=NUM
svc
`

			assert.Nil(t, err)
			assert.Equal(t, "std err\n", stderr)
			assert.Equal(t, expOut, stdout)
		})

	})
}
