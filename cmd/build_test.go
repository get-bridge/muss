package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"gerrit.instructure.com/muss/config"
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

		t.Run("build with private registry 403 with service def", func(*testing.T) {
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

			stdout, stderr, err := testCmdBuilder(newBuildCommand, []string{})

			assert.Equal(t, "exit status 1", err.Error())
			assert.Equal(t, "std err\nBuilding test\nService 'test' failed to build: error parsing HTTP 403 response body: unexpected end of JSON input: \"\"\n\nYou may need to login to myreg.docker\n", stderr)
			expOut := "Step 1/1 : FROM private.registry.docker/ns/image:tag\n"
			assert.Equal(t, expOut, stdout)
		})

		t.Run("build with private registry 403 with unknown registry", func(*testing.T) {
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
										"image": "hoge/piyo",
									},
								},
							},
						},
					},
				},
			})
			defer config.SetConfig(nil)

			stdout, stderr, err := testCmdBuilder(newBuildCommand, []string{})

			assert.Equal(t, "exit status 1", err.Error())
			assert.Equal(t, "std err\nBuilding test\nService 'test' failed to build: error parsing HTTP 403 response body: unexpected end of JSON input: \"\"\n\nYou may need to login to your docker registry\n", stderr)
			expOut := "Step 1/1 : FROM private.registry.docker/ns/image:tag\n"
			assert.Equal(t, expOut, stdout)
		})

		t.Run("build with private registry basic auth error", func(*testing.T) {
			os.Setenv("MUSS_TEST_REGISTRY_ERROR", "no-basic-auth")
			defer os.Unsetenv("MUSS_TEST_REGISTRY_ERROR")

			stdout, stderr, err := testCmdBuilder(newBuildCommand, []string{})

			assert.Equal(t, "exit status 1", err.Error())
			assert.Equal(t, "std err\nBuilding test\nService 'test' failed to build: Get https://private.registry.docker/v2/ns/image/manifests/tag: no basic auth credentials\n\nYou may need to login to private.registry.docker\n", stderr)
			expOut := "Step 1/1 : FROM private.registry.docker/ns/image:tag\n"
			assert.Equal(t, expOut, stdout)
		})
	})
}
