package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"gerrit.instructure.com/muss/config"
)

func TestPullCommand(t *testing.T) {
	withTestPath(t, func(*testing.T) {
		t.Run("all args pass through", func(*testing.T) {
			stdout, stderr, err := runTestCommand(nil, []string{
				"pull",
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

		t.Run("pull with private registry 403 without match", func(*testing.T) {
			os.Setenv("MUSS_TEST_REGISTRY_ERROR", "403")
			defer os.Unsetenv("MUSS_TEST_REGISTRY_ERROR")

			stdout, stderr, err := runTestCommand(nil, []string{"pull"})

			assert.Equal(t, "exit status 1", err.Error())
			assert.Equal(t, "std err\nERROR: for foo  error parsing HTTP 403 response body: unexpected end of JSON input: \"\"\nERROR: for test  error parsing HTTP 403 response body: unexpected end of JSON input: \"\"\nerror parsing HTTP 403 response body: unexpected end of JSON input: \"\"\nerror parsing HTTP 403 response body: unexpected end of JSON input: \"\"\n\nYou may need to login to your docker registry\n", stderr)
			expOut := ""
			assert.Equal(t, expOut, stdout)
		})

		t.Run("pull with private registry 403 with service def", func(*testing.T) {
			os.Setenv("MUSS_TEST_REGISTRY_ERROR", "403")
			defer os.Unsetenv("MUSS_TEST_REGISTRY_ERROR")

			config.SetConfig(map[string]interface{}{
				"service_definitions": []config.ServiceDef{
					config.ServiceDef{
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
			defer config.SetConfig(nil)
			cfg, _ := config.All()

			stdout, stderr, err := runTestCommand(cfg, []string{"pull"})

			assert.Equal(t, "exit status 1", err.Error())
			assert.Equal(t, "std err\nERROR: for foo  error parsing HTTP 403 response body: unexpected end of JSON input: \"\"\nERROR: for test  error parsing HTTP 403 response body: unexpected end of JSON input: \"\"\nerror parsing HTTP 403 response body: unexpected end of JSON input: \"\"\nerror parsing HTTP 403 response body: unexpected end of JSON input: \"\"\n\nYou may need to login to myreg.docker\n", stderr)
			expOut := ""
			assert.Equal(t, expOut, stdout)
		})

		t.Run("pull with private registry basic auth error", func(*testing.T) {
			os.Setenv("MUSS_TEST_REGISTRY_ERROR", "no-basic-auth")
			defer os.Unsetenv("MUSS_TEST_REGISTRY_ERROR")

			stdout, stderr, err := runTestCommand(nil, []string{"pull"})

			assert.Equal(t, "exit status 1", err.Error())
			assert.Equal(t, "std err\nERROR: for test  Get https://private.registry.docker/v2/ns/image/manifests/tag: no basic auth credentials\nGet https://private.registry.docker/v2/ns/image/manifests/tag: no basic auth credentials\n\nYou may need to login to private.registry.docker\n", stderr)
			expOut := ""
			assert.Equal(t, expOut, stdout)
		})

	})
}
