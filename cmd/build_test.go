package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildCommand(t *testing.T) {
	withTestPath(t, func(t *testing.T) {
		t.Run("all args pass through", func(t *testing.T) {
			stdout, stderr, err := runTestCommand(nil, []string{
				"build",
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

		t.Run("build with private registry 403 with service def", func(t *testing.T) {
			os.Setenv("MUSS_TEST_REGISTRY_ERROR", "403")
			defer os.Unsetenv("MUSS_TEST_REGISTRY_ERROR")

			cfg := newTestConfig(t, map[string]interface{}{
				"service_definitions": []map[string]interface{}{
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

			stdout, stderr, err := runTestCommand(cfg, []string{"build"})

			assert.Equal(t, "exit status 1", err.Error())
			assert.Equal(t, "std err\nBuilding test\nService 'test' failed to build: error parsing HTTP 403 response body: unexpected end of JSON input: \"\"\n\nYou may need to login to myreg.docker\n", stderr)
			expOut := "Step 1/1 : FROM private.registry.docker/ns/image:tag\n"
			assert.Equal(t, expOut, stdout)
		})

		t.Run("build with private registry 403 with unknown registry", func(t *testing.T) {
			os.Setenv("MUSS_TEST_REGISTRY_ERROR", "403")
			defer os.Unsetenv("MUSS_TEST_REGISTRY_ERROR")

			cfg := newTestConfig(t, map[string]interface{}{
				"service_definitions": []map[string]interface{}{
					map[string]interface{}{
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

			stdout, stderr, err := runTestCommand(cfg, []string{"build"})

			assert.Equal(t, "exit status 1", err.Error())
			assert.Equal(t, "std err\nBuilding test\nService 'test' failed to build: error parsing HTTP 403 response body: unexpected end of JSON input: \"\"\n\nYou may need to login to your docker registry\n", stderr)
			expOut := "Step 1/1 : FROM private.registry.docker/ns/image:tag\n"
			assert.Equal(t, expOut, stdout)
		})

		t.Run("build with private registry basic auth error", func(t *testing.T) {
			os.Setenv("MUSS_TEST_REGISTRY_ERROR", "no-basic-auth")
			defer os.Unsetenv("MUSS_TEST_REGISTRY_ERROR")

			stdout, stderr, err := runTestCommand(nil, []string{"build"})

			assert.Equal(t, "exit status 1", err.Error())
			assert.Equal(t, "std err\nBuilding test\nService 'test' failed to build: Get https://private.registry.docker/v2/ns/image/manifests/tag: no basic auth credentials\n\nYou may need to login to private.registry.docker\n", stderr)
			expOut := "Step 1/1 : FROM private.registry.docker/ns/image:tag\n"
			assert.Equal(t, expOut, stdout)
		})
	})
}
