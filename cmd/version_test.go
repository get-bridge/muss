package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionCommand(t *testing.T) {
	Version = "testing"
	withTestPath(t, func(*testing.T) {
		t.Run("version info", func(*testing.T) {

			stdout, stderr, err := testCmdBuilder(newVersionCommand, []string{})

			expOut := `muss testing
docker-compose fake
docker
version
--format
docker client {{ .Client.Version }}{{ "\n" }}docker server {{ .Server.Version }}
`

			assert.Nil(t, err)
			assert.Equal(t, "", stderr)
			assert.Equal(t, expOut, stdout)
		})

		t.Run("errors", func(*testing.T) {
			os.Setenv("MUSS_TEST_DOCKER_VERSION_ERROR", "1")
			defer os.Unsetenv("MUSS_TEST_DOCKER_VERSION_ERROR")

			stdout, stderr, err := testCmdBuilder(newVersionCommand, []string{})

			expOut := `muss testing
docker-compose fake
error getting docker version: exit status 2
some errors
`

			assert.Nil(t, err)
			assert.Equal(t, "", stderr)
			assert.Equal(t, expOut, stdout)
		})
	})
}
