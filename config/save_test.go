package config

import (
	"io/ioutil"
	"os"
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

	dir := tempdir(t)
	os.Chdir(dir)
	defer func() {
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
				},
			},
		}
		cfg := map[string]interface{}{
			"service_definitions": []ServiceDef{
				map[string]interface{}{
					"name": "app",
					"configs": map[string]interface{}{
						"sole": exp,
					},
				},
			},
		}
		ioutil.WriteFile(ProjectFile, yamlDump(cfg), 0644)

		assert.Nil(t, project, "Project config not yet loaded") // prove that Save will load it.
		Save()
		assert.NotNil(t, project, "Save loads project config first")

		if written, err := ioutil.ReadFile(DockerComposeFile); err != nil {
			t.Fatalf("failed to open generated file: %s\n", err)
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
	})
}
