package config

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigSave(t *testing.T) {
	withTempDir(t, func(tmpdir string) {
		UserFile = "muss.test.yaml"
		defer func() {
			UserFile = ""
		}()
		os.Unsetenv("MUSS_TEST_VAR")
		os.Unsetenv("MUSS_SECRET_TEST")
		os.Unsetenv("MUSS_SECRET_TEST_TWO")
		os.Setenv("MUSS_TEST_PASSPHRASE", "phrasey")

		t.Run("no config", func(t *testing.T) {
			SetConfig(nil)
			generateFiles(nil)
			// no errors
		})

		t.Run("config save", func(t *testing.T) {

			exp := map[string]interface{}{
				"version": "3.6",
				"volumes": map[string]interface{}{},
				"services": map[string]interface{}{
					"app": map[string]interface{}{
						"image": "alpine",
						"environment": map[string]interface{}{
							"FOO": "bar",
						},
						"volumes": []interface{}{
							"./foo:/bar",
							"./pre-existing.file:/some/file",
							map[string]interface{}{
								"type":   "bind",
								"source": "${MUSS_TEST_VAR:-~/vol}/file",
								"target": "/filevol",
							},
						},
					},
				},
			}
			cfg := map[string]interface{}{
				"secret_passphrase": "$MUSS_TEST_PASSPHRASE",
				"service_definitions": []ServiceDef{
					map[string]interface{}{
						"name": "app",
						"configs": map[string]interface{}{
							"sole": map[string]interface{}{
								"version": "3.6",
								"volumes": map[string]interface{}{},
								"services": map[string]interface{}{
									"app": map[string]interface{}{
										"image": "alpine",
										"environment": map[string]interface{}{
											"FOO": "bar",
										},
										"volumes": []interface{}{
											"./foo:/bar",
											"./pre-existing.file:/some/file",
											map[string]interface{}{
												"type":   "bind",
												"source": "${MUSS_TEST_VAR:-~/vol}/file",
												"target": "/filevol",
												"file":   true,
											},
										},
									},
								},
								"secrets": map[string]interface{}{
									"MUSS_SECRET_TEST": map[string]interface{}{
										"exec": []string{"echo", "hello"},
									},
									"MUSS_SECRET_TEST_TWO": map[string]interface{}{
										"exec": []string{"echo", "goodbye"},
									},
								},
							},
						},
					},
				},
			}
			ioutil.WriteFile(ProjectFile, yamlDump(cfg), 0644)

			assertNotExist(t, "./foo")
			assertNotExist(t, "./test-home/vol/file")
			assertNotExist(t, "./pre-existing.file")
			touch("./pre-existing.file")

			assert.Nil(t, project, "Project config not yet loaded") // prove that Save will load it.
			Save()
			assert.NotNil(t, project, "Save loads project config first")

			if written, err := ioutil.ReadFile(DockerComposeFile); err != nil {
				t.Fatalf("failed to open generated file: %s", err)
			} else {

				assert.Contains(t,
					string(written),
					"# To add new service definition files edit "+ProjectFile+".",
					"contains generated comments",
				)

				assert.Contains(t,
					string(written),
					"# To configure the services you want to use edit "+UserFile+".",
					"contains user file comments",
				)

				parsed, err := parseYaml(written)
				if err != nil {
					t.Fatalf("failed to parse yaml: %s\n", err)
				}
				assert.EqualValues(t, exp, parsed, "Generated docker-compose yaml")
			}

			assert.DirExists(t, "./foo", "plain file")
			assert.FileExists(t, "./test-home/vol/file", "plain file")
			assert.FileExists(t, "./pre-existing.file", "still a file")

			assert.Equal(t, "hello", os.Getenv("MUSS_SECRET_TEST"), "loaded secret")
			assert.Equal(t, "goodbye", os.Getenv("MUSS_SECRET_TEST_TWO"), "loaded second secret")
		})

		t.Run("ensureFile", func(t *testing.T) {
			if err := os.MkdirAll("foo/bar/baz", 0777); err != nil {
				t.Fatalf("failed to mkdir: %s", err)
			}
			if err := ensureFile("foo/bar"); err == nil {
				t.Fatal("expected error, got none")
			} else {
				assert.Equal(t, "remove foo/bar: directory not empty", err.Error())
			}

			err := ensureFile("foo/bar/baz")
			assert.Nil(t, err, "no error changing dir to file")
			assert.FileExists(t, "foo/bar/baz")

			again := ensureFile("foo/bar/baz")
			assert.Nil(t, again, "no error when already a file")
			assert.FileExists(t, "foo/bar/baz")
		})
	})
}
