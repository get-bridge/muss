package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionCommand(t *testing.T) {
	Version = "testing"
	withTestPath(t, func(t *testing.T) {
		t.Run("version info", func(t *testing.T) {

			stdout, stderr, err := runTestCommand(nil, []string{"version"})

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

		t.Run("short", func(t *testing.T) {
			stdout, stderr, err := runTestCommand(nil, []string{"version", "--short"})

			expOut := "testing\n"

			assert.Nil(t, err)
			assert.Equal(t, "", stderr)
			assert.Equal(t, expOut, stdout)
		})

		t.Run("errors", func(t *testing.T) {
			os.Setenv("MUSS_TEST_DOCKER_VERSION_ERROR", "1")
			defer os.Unsetenv("MUSS_TEST_DOCKER_VERSION_ERROR")

			stdout, stderr, err := runTestCommand(nil, []string{"version"})

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
