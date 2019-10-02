package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildCommand(t *testing.T) {
	withTestPath(t, func(*testing.T) {
		t.Run("all args pass through", func(*testing.T) {
			stdout, stderr, err := testCmdBuilder(newBuildCommand, []string{
				"--compress",
				"--force-rm",
				"--no-cache",
				"--pull",
				"-m", ":nomemory:",
				"--build-arg", "k=some val",
				"svc1",
				"svc2",
			})

			expOut := `docker-compose
build
--compress
--force-rm
--no-cache
--pull
-m
:nomemory:
--build-arg
k=some val
svc1
svc2
`

			assert.Nil(t, err)
			assert.Equal(t, "std err\n", stderr)
			assert.Equal(t, expOut, stdout)
		})
	})
}
