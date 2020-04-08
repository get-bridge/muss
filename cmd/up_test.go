package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"gerrit.instructure.com/muss/term"
)

func TestUpCommand(t *testing.T) {
	withTestPath(t, func(t *testing.T) {
		t.Run("all args pass through", func(t *testing.T) {
			stdout, stderr, err := runTestCommand(nil, []string{
				"up",
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

		t.Run("stop all", func(t *testing.T) {
			stdout, stderr, err := runTestCommand(nil, []string{"up"})

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

		t.Run("stop selected", func(t *testing.T) {
			stdout, stderr, err := runTestCommand(nil, []string{"up", "--no-status", "hoge", "piyo"})

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

		t.Run("up without starting in foreground", func(t *testing.T) {
			os.Setenv("MUSS_TEST_UP_LOGS", "1")
			defer os.Unsetenv("MUSS_TEST_UP_LOGS")

			args := []string{"-d", "--no-start"}

			cfg := newTestConfig(t, nil)
			for _, arg := range args {
				stdout, stderr, err := runTestCommand(cfg, []string{"up", arg})

				expOut := "log\n"

				assert.Nil(t, err)
				assert.Equal(t, "", stderr)
				assert.Equal(t, expOut, stdout)
			}
		})

		t.Run("up --no-status", func(t *testing.T) {
			os.Setenv("MUSS_TEST_UP_LOGS", "1")
			defer os.Unsetenv("MUSS_TEST_UP_LOGS")

			stdout, stderr, err := runTestCommand(nil, []string{"up", "--no-status", "hoge", "piyo"})

			expOut := "log\ndocker-compose\nstop\nhoge\npiyo\n"

			assert.Nil(t, err)
			assert.Equal(t, "std err\n", stderr)
			assert.Equal(t, expOut, stdout)
		})

		t.Run("up without muss.yaml", func(t *testing.T) {
			os.Setenv("MUSS_TEST_UP_LOGS", "1")
			defer os.Unsetenv("MUSS_TEST_UP_LOGS")

			stdout, stderr, err := runTestCommand(nil, []string{"up"})

			expOut := term.AnsiEraseToEnd +
				term.AnsiReset + "# muss" + term.AnsiReset + term.AnsiStart +
				term.AnsiEraseToEnd + "log\n" +
				term.AnsiReset + "# muss" + term.AnsiReset + term.AnsiStart +
				"docker-compose\nstop\n"

			assert.Nil(t, err)
			assert.Equal(t, "std err\n", stderr)
			assert.Equal(t, expOut, stdout)
		})

		t.Run("up with status", func(t *testing.T) {
			os.Setenv("MUSS_TEST_UP_LOGS", "1")
			defer os.Unsetenv("MUSS_TEST_UP_LOGS")

			cfg := newTestConfig(t, map[string]interface{}{
				"status": map[string]interface{}{
					"exec":        []string{"../testdata/bin/status"},
					"interval":    "1.1s",
					"line_format": "# %s",
				},
			})

			stdout, stderr, err := runTestCommand(cfg, []string{"up"})

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

		t.Run("up with multi line status", func(t *testing.T) {
			os.Setenv("MUSS_TEST_UP_LOGS", "1")
			defer os.Unsetenv("MUSS_TEST_UP_LOGS")

			cfg := newTestConfig(t, map[string]interface{}{
				"status": map[string]interface{}{
					"exec":        []string{"../testdata/bin/status", "prefix"},
					"interval":    "1.1s",
					"line_format": "# %s",
				},
			})

			stdout, stderr, err := runTestCommand(cfg, []string{"up"})

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

		t.Run("up with status and private registry 403", func(t *testing.T) {
			os.Setenv("MUSS_TEST_REGISTRY_ERROR", "403")
			defer os.Unsetenv("MUSS_TEST_REGISTRY_ERROR")

			cfg := newTestConfig(t, map[string]interface{}{
				"status": map[string]interface{}{
					"exec":        []string{"../testdata/bin/status", "prefix"},
					"interval":    "1.1s",
					"line_format": "# %s",
				},
			})

			stdout, stderr, err := runTestCommand(cfg, []string{"up"})

			assert.Equal(t, "exit status 1", err.Error())
			assert.Equal(t, "std err\nPulling test (private.registry.docker/ns/image:tag)...\nerror parsing HTTP 403 response body: unexpected end of JSON input: \"\"\n\nYou may need to login to private.registry.docker\n", stderr)
			expOut := term.AnsiEraseToEnd +
				term.AnsiReset + "# muss" + term.AnsiReset + term.AnsiStart
			assert.Equal(t, expOut, stdout)
		})

		t.Run("up without status and private registry basic-auth error", func(t *testing.T) {
			os.Setenv("MUSS_TEST_REGISTRY_ERROR", "no-basic-auth")
			defer os.Unsetenv("MUSS_TEST_REGISTRY_ERROR")

			stdout, stderr, err := runTestCommand(nil, []string{"up", "--no-status"})

			assert.Equal(t, "exit status 1", err.Error())
			assert.Equal(t, "std err\nPulling test (private.registry.docker/ns/image:tag)...\nGet https://private.registry.docker/v2/ns/image/manifests/tag: no basic auth credentials\n\nYou may need to login to private.registry.docker\n", stderr)
			expOut := ""
			assert.Equal(t, expOut, stdout)
		})

		t.Run("up with private registry 403 on build", func(t *testing.T) {
			os.Setenv("MUSS_TEST_REGISTRY_ERROR", "build-403")
			defer os.Unsetenv("MUSS_TEST_REGISTRY_ERROR")

			cfg := newTestConfig(t, map[string]interface{}{
				"module_definitions": []map[string]interface{}{
					map[string]interface{}{
						"name": "app",
						"configs": map[string]interface{}{
							"sole": map[string]interface{}{
								"services": map[string]interface{}{
									"test": map[string]interface{}{
										"image": "myreg.docker/hoge/piyo",
									},
								},
							},
						},
					},
				},
			})

			stdout, stderr, err := runTestCommand(cfg, []string{"up", "--no-status"})

			assert.Equal(t, "exit status 1", err.Error())
			assert.Equal(t, "std err\nBuilding test\nService 'test' failed to build: error parsing HTTP 403 response body: unexpected end of JSON input: \"\"\n\nYou may need to login to myreg.docker\n", stderr)
			expOut := "Step 1/1 : FROM private.registry.docker/ns/image:tag\n"
			assert.Equal(t, expOut, stdout)
		})

		t.Run("up with private registry cred-helper stack trach", func(t *testing.T) {
			os.Setenv("MUSS_TEST_REGISTRY_ERROR", "cred-helper")
			defer os.Unsetenv("MUSS_TEST_REGISTRY_ERROR")

			stdout, stderr, err := runTestCommand(nil, []string{"up", "--no-status"})

			assert.Equal(t, "exit status 1", err.Error())

			lines := strings.SplitAfter(stderr, "\n")
			length := len(lines)
			if length < 5 {
				t.Fatalf("Not enough stderr lines:\n%s", stderr)
			}
			assert.Equal(t, "std err\n", lines[0])
			assert.Equal(t, "[pid] Failed to execute script docker-compose\n", lines[1])
			// ...
			assert.Equal(t, `docker.errors.DockerException: Credentials store error: StoreError('Credentials store docker-credential-desktop exited with "No stored credential for private.registry.docker".')`+"\n", lines[length-4])
			assert.Equal(t, "\n", lines[length-3], "blank")
			assert.Equal(t, "You may need to login to private.registry.docker\n", lines[length-2])
			assert.Equal(t, "", lines[length-1], "ended with newline")

			expOut := ""
			assert.Equal(t, expOut, stdout)
		})

	})
}
