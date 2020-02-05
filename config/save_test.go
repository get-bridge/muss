package config

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigSave(t *testing.T) {
	withTempDir(t, func(tmpdir string) {
		os.Setenv("MUSS_USER_FILE", "muss.test.yaml")
		defer os.Unsetenv("MUSS_USER_FILE")
		os.Unsetenv("COMPOSE_FILE")
		os.Unsetenv("COMPOSE_PATH_SEPARATOR")
		os.Unsetenv("MUSS_TEST_VAR")
		os.Unsetenv("MUSS_SECRET_TEST")
		os.Unsetenv("MUSS_SECRET_TEST_TWO")
		os.Setenv("MUSS_TEST_PASSPHRASE", "phrasey")

		t.Run("no config", func(t *testing.T) {
			SetConfig(nil)
			generateFiles(nil)
			// no errors

			// no compose file
			assertNotExist(t, "docker-compose.yml")

			SetConfig(map[string]interface{}{
				"status": map[string]interface{}{
					"exec": []string{"echo", "hi"},
				},
			})

			cfg, err := All()
			if err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, cfg.Status.Exec, []string{"echo", "hi"}, "has config")
			generateFiles(cfg)
			// still no compose file
			assertNotExist(t, "docker-compose.yml")
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
							// Test that we ignore permission errors.
							"/muss-test-dir:/muss-test-dir",
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
											"/muss-test-dir:/muss-test-dir",
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

			if yaml, err := yamlDump(cfg); err != nil {
				t.Fatal(err)
			} else {
				ioutil.WriteFile(ProjectFile, yaml, 0644)
			}

			assertNotExist(t, "./foo")
			assertNotExist(t, "./test-home/vol/file")
			assertNotExist(t, "./pre-existing.file")
			touch("./pre-existing.file")

			var stderr bytes.Buffer
			SetStderr(&stderr)

			SetConfig(nil)
			assert.Nil(t, project, "Project config not yet loaded") // prove that Save will load it.
			Save()
			assert.NotNil(t, project, "Save loads project config first")

			if written, err := ioutil.ReadFile("docker-compose.yml"); err != nil {
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

			assert.Equal(t, "", stderr.String(), "no warnings")
		})

		setAndSaveTiny := func(cfg map[string]interface{}, expTarget string) string {
			exp := map[string]interface{}{
				"version": "3.4",
			}
			cfg["service_definitions"] = []ServiceDef{
				map[string]interface{}{
					"name": "app",
					"configs": map[string]interface{}{
						"sole": map[string]interface{}{
							"version": "3.4",
						},
					},
				},
			}
			SetConfig(cfg)

			var stderr bytes.Buffer
			SetStderr(&stderr)

			Save()

			if written, err := ioutil.ReadFile(expTarget); err != nil {
				t.Fatalf("failed to open generated file: %s", err)
			} else {

				assert.Contains(t,
					string(written),
					"# To add new service definition files edit "+ProjectFile+".",
					"contains generated comments",
				)

				parsed, err := parseYaml(written)
				if err != nil {
					t.Fatalf("failed to parse yaml: %s\n", err)
				}
				assert.EqualValues(t, exp, parsed, "Generated docker-compose yaml")
			}

			return stderr.String()
		}

		t.Run("COMPOSE_FILE is set", func(t *testing.T) {
			os.Setenv("COMPOSE_FILE", "dc-test.yml")
			defer os.Unsetenv("COMPOSE_FILE")

			warning := setAndSaveTiny(map[string]interface{}{}, "docker-compose.yml")

			assert.Equal(t, "COMPOSE_FILE is set but does not contain muss target 'docker-compose.yml'.\n", warning, "warning about COMPOSE_FILE")
		})

		t.Run("compose_file config", func(t *testing.T) {
			warning := setAndSaveTiny(map[string]interface{}{"compose_file": "dc.test.yml"}, "dc.test.yml")

			assert.Equal(t, "", warning, "no warnings")
		})

		t.Run("compose_file config with COMPOSE_FILE", func(t *testing.T) {
			os.Setenv("COMPOSE_FILE", "dc-other.yml")
			defer os.Unsetenv("COMPOSE_FILE")

			warning := setAndSaveTiny(map[string]interface{}{"compose_file": "dc.test.yml"}, "dc.test.yml")

			assert.Equal(t, "COMPOSE_FILE is set but does not contain muss target 'dc.test.yml'.\n", warning, "warning about COMPOSE_FILE")
		})

		t.Run("COMPOSE_FILE warnings", func(t *testing.T) {
			defer os.Unsetenv("COMPOSE_FILE")
			defer os.Unsetenv("COMPOSE_PATH_SEPARATOR")
			os.Unsetenv("COMPOSE_FILE")

			var stderr bytes.Buffer
			SetStderr(&stderr)
			SetConfig(map[string]interface{}{})

			warning := func() string {
				stderr.Reset()
				project.checkComposeFileVar()
				return stderr.String()
			}

			assert.Equal(t, "", warning(), "no warning when unset")

			os.Setenv("COMPOSE_FILE", "foo.yml:./docker-compose.yml:bar.yml")
			assert.Equal(t, "", warning(), "no warning when included")

			os.Setenv("COMPOSE_PATH_SEPARATOR", ";")
			os.Setenv("COMPOSE_FILE", "foo.yml;./docker-compose.yml;bar.yml")
			assert.Equal(t, "", warning(), "no warning when included (alternate separator)")

			os.Unsetenv("COMPOSE_PATH_SEPARATOR")
			assert.Equal(t, "COMPOSE_FILE is set but does not contain muss target 'docker-compose.yml'.\n", warning(), "warning when not found")
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
