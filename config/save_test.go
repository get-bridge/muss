package config

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/instructure-bridge/muss/testutil"
)

func TestConfigSave(t *testing.T) {
	testutil.WithTempDir(t, func(tmpdir string) {
		os.Setenv("MUSS_USER_FILE", "muss.test.yaml")
		defer os.Unsetenv("MUSS_USER_FILE")
		os.Unsetenv("COMPOSE_FILE")
		os.Unsetenv("COMPOSE_PATH_SEPARATOR")
		os.Unsetenv("MUSS_TEST_VAR")
		os.Unsetenv("MUSS_SECRET_TEST")
		os.Unsetenv("MUSS_SECRET_TEST_TWO")
		os.Setenv("MUSS_TEST_PASSPHRASE", "phrasey")

		t.Run("no config", func(t *testing.T) {
			cfg := newTestConfig(t, nil)
			err := cfg.Save()
			assert.Nil(t, err)

			// no compose file
			testutil.NoFileExists(t, "docker-compose.yml")

			cfg = newTestConfig(t, map[string]interface{}{
				"status": map[string]interface{}{
					"exec": []string{"echo", "hi"},
				},
			})

			assert.Equal(t, cfg.Status.Exec, []string{"echo", "hi"}, "has config")
			err = cfg.Save()
			assert.Nil(t, err)
			// still no compose file
			testutil.NoFileExists(t, "docker-compose.yml")
		})

		t.Run("config save", func(t *testing.T) {
			os.Unsetenv("MUSS_FILE")
			os.Unsetenv("MUSS_USER_FILE")

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
			cfgMap := map[string]interface{}{
				"secret_passphrase": "$MUSS_TEST_PASSPHRASE",
				"module_definitions": []map[string]interface{}{
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

			if yaml, err := yamlDump(cfgMap); err != nil {
				t.Fatal(err)
			} else {
				testutil.WriteFile(t, defaultProjectFile, string(yaml))
			}

			testutil.NoFileExists(t, "./foo")
			testutil.NoFileExists(t, "./test-home/vol/file")
			testutil.NoFileExists(t, "./pre-existing.file")
			touch("./pre-existing.file")

			cfg, _ := NewConfigFromDefaultFile()
			err := cfg.Save()

			assert.Equal(t, "$MUSS_TEST_PASSPHRASE", cfg.SecretPassphrase, "loaded")
			assert.Nil(t, err, "no errors")

			if written, err := ioutil.ReadFile("docker-compose.yml"); err != nil {
				t.Fatalf("failed to open generated file: %s", err)
			} else {

				assert.Contains(t,
					string(written),
					"# To add new module definition files edit muss.yaml.",
					"contains generated comments",
				)

				assert.Contains(t,
					string(written),
					"# To configure the modules you want to use edit muss.user.yaml.",
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

			assert.Equal(t, 0, len(cfg.Warnings), "no warnings")
		})

		setAndSaveTiny := func(cfgMap map[string]interface{}, expTarget string) string {
			exp := map[string]interface{}{
				"version": "3.4",
			}
			cfgMap["module_definitions"] = []map[string]interface{}{
				map[string]interface{}{
					"name": "app",
					"configs": map[string]interface{}{
						"sole": map[string]interface{}{
							"version": "3.4",
						},
					},
				},
			}
			cfg := newTestConfig(t, cfgMap)
			cfg.ProjectFile = "test.file"

			cfg.Save()

			if written, err := ioutil.ReadFile(expTarget); err != nil {
				t.Fatalf("failed to open generated file: %s", err)
			} else {

				assert.Contains(t,
					string(written),
					"# To add new module definition files edit test.file.",
					"contains generated comments",
				)

				parsed, err := parseYaml(written)
				if err != nil {
					t.Fatalf("failed to parse yaml: %s\n", err)
				}
				assert.EqualValues(t, exp, parsed, "Generated docker-compose yaml")
			}

			return strings.Join(cfg.Warnings, "\n")
		}

		t.Run("COMPOSE_FILE is set", func(t *testing.T) {
			os.Setenv("COMPOSE_FILE", "dc-test.yml")
			defer os.Unsetenv("COMPOSE_FILE")

			warning := setAndSaveTiny(map[string]interface{}{}, "docker-compose.yml")

			assert.Equal(t, "COMPOSE_FILE is set but does not contain muss target 'docker-compose.yml'.", warning, "warning about COMPOSE_FILE")
		})

		t.Run("compose_file config", func(t *testing.T) {
			warning := setAndSaveTiny(map[string]interface{}{"compose_file": "dc.test.yml"}, "dc.test.yml")

			assert.Equal(t, "", warning, "no warnings")
		})

		t.Run("compose_file config with COMPOSE_FILE", func(t *testing.T) {
			os.Setenv("COMPOSE_FILE", "dc-other.yml")
			defer os.Unsetenv("COMPOSE_FILE")

			warning := setAndSaveTiny(map[string]interface{}{"compose_file": "dc.test.yml"}, "dc.test.yml")

			assert.Equal(t, "COMPOSE_FILE is set but does not contain muss target 'dc.test.yml'.", warning, "warning about COMPOSE_FILE")
		})

		t.Run("COMPOSE_FILE warnings", func(t *testing.T) {
			defer os.Unsetenv("COMPOSE_FILE")
			defer os.Unsetenv("COMPOSE_PATH_SEPARATOR")
			os.Unsetenv("COMPOSE_FILE")

			cfg := newTestConfig(t, nil)

			warning := func() string {
				cfg.Warnings = nil
				cfg.checkComposeFileVar()
				return strings.Join(cfg.Warnings, "\n")
			}

			assert.Equal(t, "", warning(), "no warning when unset")

			os.Setenv("COMPOSE_FILE", "foo.yml:./docker-compose.yml:bar.yml")
			assert.Equal(t, "", warning(), "no warning when included")

			os.Setenv("COMPOSE_PATH_SEPARATOR", ";")
			os.Setenv("COMPOSE_FILE", "foo.yml;./docker-compose.yml;bar.yml")
			assert.Equal(t, "", warning(), "no warning when included (alternate separator)")

			os.Unsetenv("COMPOSE_PATH_SEPARATOR")
			assert.Equal(t, "COMPOSE_FILE is set but does not contain muss target 'docker-compose.yml'.", warning(), "warning when not found")
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
