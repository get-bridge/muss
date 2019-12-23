package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"gerrit.instructure.com/muss/config"
	"gerrit.instructure.com/muss/term"
)

func TestUpCommand(t *testing.T) {
	withTestPath(t, func(*testing.T) {
		t.Run("all args pass through", func(*testing.T) {
			stdout, stderr, err := testCmdBuilder(newUpCommand, []string{
				"--no-status",
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
--detach
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
--timeout=4
--renew-anon-volumes
--remove-orphans
--exit-code-from=svc
--scale=SERVICE=NUM
svc
`

			assert.Nil(t, err)
			assert.Equal(t, "std err\n", stderr)
			assert.Equal(t, expOut, stdout)
		})

		t.Run("stop all", func(*testing.T) {
			stdout, stderr, err := testCmdBuilder(newUpCommand, []string{})

			expOut := term.AnsiEraseToEnd +
				term.AnsiReset + "# muss" + term.AnsiReset + term.AnsiStart +
				term.AnsiEraseToEnd + "docker-compose\n" +
				term.AnsiReset + "# muss" + term.AnsiReset + term.AnsiStart +
				term.AnsiEraseToEnd + "up\n" +
				term.AnsiReset + "# muss" + term.AnsiReset + term.AnsiStart +
				"docker-compose\nstop\n"

			assert.Nil(t, err)
			assert.Equal(t, "std err\nstd err\n", stderr)
			assert.Equal(t, expOut, stdout)
		})

		t.Run("stop selected", func(*testing.T) {
			stdout, stderr, err := testCmdBuilder(newUpCommand, []string{"--no-status", "hoge", "piyo"})

			expOut := `docker-compose
up
hoge
piyo
docker-compose
stop
hoge
piyo
`

			assert.Nil(t, err)
			assert.Equal(t, "std err\nstd err\n", stderr)
			assert.Equal(t, expOut, stdout)
		})

		t.Run("up without starting in foreground", func(*testing.T) {
			os.Setenv("MUSS_TEST_UP_LOGS", "1")
			defer os.Unsetenv("MUSS_TEST_UP_LOGS")

			config.SetConfig(nil)

			args := []string{"-d", "--no-start"}

			for _, arg := range args {
				stdout, stderr, err := testCmdBuilder(newUpCommand, []string{arg})

				expOut := "log\n"

				assert.Nil(t, err)
				assert.Equal(t, "", stderr)
				assert.Equal(t, expOut, stdout)
			}
		})

		t.Run("up --no-status", func(*testing.T) {
			os.Setenv("MUSS_TEST_UP_LOGS", "1")
			defer os.Unsetenv("MUSS_TEST_UP_LOGS")

			config.SetConfig(nil)

			stdout, stderr, err := testCmdBuilder(newUpCommand, []string{"--no-status", "hoge", "piyo"})

			expOut := "log\ndocker-compose\nstop\nhoge\npiyo\n"

			assert.Nil(t, err)
			assert.Equal(t, "std err\n", stderr)
			assert.Equal(t, expOut, stdout)
		})

		t.Run("up without muss.yaml", func(*testing.T) {
			os.Setenv("MUSS_TEST_UP_LOGS", "1")
			defer os.Unsetenv("MUSS_TEST_UP_LOGS")

			config.SetConfig(nil)

			stdout, stderr, err := testCmdBuilder(newUpCommand, []string{})

			expOut := term.AnsiEraseToEnd +
				term.AnsiReset + "# muss" + term.AnsiReset + term.AnsiStart +
				term.AnsiEraseToEnd + "log\n" +
				term.AnsiReset + "# muss" + term.AnsiReset + term.AnsiStart +
				"docker-compose\nstop\n"

			assert.Nil(t, err)
			assert.Equal(t, "std err\n", stderr)
			assert.Equal(t, expOut, stdout)
		})

		t.Run("up with status", func(*testing.T) {
			os.Setenv("MUSS_TEST_UP_LOGS", "1")
			defer os.Unsetenv("MUSS_TEST_UP_LOGS")

			config.SetConfig(map[string]interface{}{
				"status": map[string]interface{}{
					"exec":        []string{"../testdata/bin/status"},
					"interval":    "1.1s",
					"line_format": "# %s",
				},
			})

			stdout, stderr, err := testCmdBuilder(newUpCommand, []string{})

			expOut := term.AnsiEraseToEnd +
				term.AnsiReset + "# muss" + term.AnsiReset + term.AnsiStart +
				term.AnsiEraseToEnd + "log\n" +
				term.AnsiReset + "# muss" + term.AnsiReset + term.AnsiStart +
				term.AnsiEraseToEnd + term.AnsiReset + "# ok!" + term.AnsiReset + term.AnsiStart +
				"docker-compose\nstop\n"

			assert.Nil(t, err)
			assert.Equal(t, "std err\n", stderr)
			assert.Equal(t, expOut, stdout)
		})

		t.Run("up with multi line status", func(*testing.T) {
			os.Setenv("MUSS_TEST_UP_LOGS", "1")
			defer os.Unsetenv("MUSS_TEST_UP_LOGS")

			config.SetConfig(map[string]interface{}{
				"status": map[string]interface{}{
					"exec":        []string{"../testdata/bin/status", "prefix"},
					"interval":    "1.1s",
					"line_format": "# %s",
				},
			})

			stdout, stderr, err := testCmdBuilder(newUpCommand, []string{})

			status := "# prefix\n# ok!"
			expOut := term.AnsiEraseToEnd +
				term.AnsiReset + "# muss" + term.AnsiReset + term.AnsiStart +
				term.AnsiEraseToEnd + "log\n" +
				term.AnsiReset + "# muss" + term.AnsiReset + term.AnsiStart +
				term.AnsiEraseToEnd + term.AnsiReset + status + term.AnsiReset + term.AnsiStart + term.AnsiUp +
				"docker-compose\nstop\n"

			assert.Nil(t, err)
			assert.Equal(t, "std err\n", stderr)
			assert.Equal(t, expOut, stdout)
		})

	})
}
