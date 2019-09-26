package config

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func tempdir(t *testing.T) string {
	tmp, err := ioutil.TempDir("", "muss-test")
	if err != nil {
		t.Fatalf("error creating temp dir: %s\n", err)
	}
	return tmp
}

func TestConfigSave(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current dir: %s\n", err)
	}

	UserFile = "muss.test.yaml"

	os.Unsetenv("MUSS_TEST_VAR")
	home := os.Getenv("HOME")
	dir := tempdir(t)
	os.Setenv("HOME", path.Join(dir, "test-home"))
	os.Chdir(dir)
	defer func() {
		os.Setenv("HOME", home)
		os.Chdir(cwd)
		os.RemoveAll(dir)
		UserFile = ""
	}()

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
}

func assertNotExist(t *testing.T, path string) {
	if _, err := os.Lstat(path); err == nil {
		t.Fatalf("expected '%s' not to exist but it does", path)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat error: %s", err)
	}
}
