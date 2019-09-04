package config

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

// NOTE: For YAML:
// - don't let any tabs get in
// - file paths are relative to this test file's parent dir

func readFile(t *testing.T, path string) string {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("error reading '%s': %s", path, err)
	}
	return string(content)
}

func assertComposed(t *testing.T, config, exp string, msgAndArgs ...interface{}) {
	var parsedExp, parsedConfig map[string]interface{}
	var err error

	parsedExp, err = parseYaml([]byte(exp))
	if err != nil {
		t.Fatalf("Error parsing exp yaml: %s", err)
	}

	parsedConfig, err = parseYaml([]byte(config))
	if err != nil {
		t.Fatalf("Error parsing config yaml: %s", err)
	}

	parsedConfig, err = prepare(parsedConfig)
	if err != nil {
		t.Fatalf("Error parsing config: %s", err)
	}

	assert.Equal(t,
		parsedExp,
		DockerComposeConfig(parsedConfig),
		msgAndArgs...)
}

func TestDockerComposeConfig(t *testing.T) {
	serviceFiles := `
service_files:
  - ../testdata/app.yml
  - ../testdata/microservice.yml
  - ../testdata/store.yml
`
	preferRepo := `
default_service_preference: [repo, registry]
`

	preferRegistry := `
default_service_preference: [registry, repo]
`

	expRepo := readFile(t, "../testdata/expectations/repo.yml")
	expRegistry := readFile(t, "../testdata/expectations/registry.yml")

	t.Run("repo preference", func(t *testing.T) {
		config := preferRepo + serviceFiles
		assertComposed(t, config, expRepo, "Config with repo preference")
		// TODO: user pref
	})

	t.Run("registry preference", func(t *testing.T) {
		config := preferRegistry + serviceFiles
		assertComposed(t, config, expRegistry, "Config with registry preference")
		// TODO: user pref
	})
}
