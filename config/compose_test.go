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

func parseAndCompose(config string) (parsed ProjectConfig, err error) {
	parsed, err = parseYaml([]byte(config))
	if err != nil {
		return
	}

	parsed, err = prepare(parsed)
	if err != nil {
		return
	}

	parsed, err = DockerComposeConfig(parsed)

	return
}

func assertConfigError(t *testing.T, config, expErr string, msgAndArgs ...interface{}) {
	_, err := parseAndCompose(config)
	if err == nil {
		t.Fatal("expected error, found nil")
	}
	assert.Contains(t, err.Error(), expErr, msgAndArgs...)
}

func assertComposed(t *testing.T, config, exp string, msgAndArgs ...interface{}) {
	var parsedExp, dc map[string]interface{}
	var err error

	parsedExp, err = parseYaml([]byte(exp))
	if err != nil {
		t.Fatalf("Error parsing exp yaml: %s", err)
	}

	dc, err = parseAndCompose(config)
	if err != nil {
		t.Fatalf("Error parsing config: %s", err)
	}

	assert.Equal(t,
		parsedExp,
		dc,
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

	userRepo := `
user: {service_preference: [repo, registry]}
`

	userRegistry := `
user: {service_preference: [registry, repo]}
`

	expRepo := readFile(t, "../testdata/expectations/repo.yml")
	expRegistry := readFile(t, "../testdata/expectations/registry.yml")

	t.Run("repo preference", func(t *testing.T) {
		config := preferRepo + serviceFiles
		assertComposed(t, config, expRepo, "Config with repo preference")

		assertComposed(t, (config + userRegistry), expRegistry, "user preference overrides")
	})

	t.Run("registry preference", func(t *testing.T) {
		config := preferRegistry + serviceFiles
		assertComposed(t, config, expRegistry, "Config with registry preference")

		assertComposed(t, (config + userRepo), expRepo, "user preference overrides")
	})

	t.Run("user custom service config", func(t *testing.T) {
		config := preferRepo + serviceFiles + `
user_file: ../testdata/user-registry-ms-repo.yml
`

		exp := readFile(t, "../testdata/expectations/user-registry-ms-repo.yml")

		assertComposed(t, config, exp, "user preference overrides orders")
	})

	t.Run("user override", func(t *testing.T) {
		config := preferRegistry + serviceFiles + `
user:
  override:
    version: '3.5'
    volumes: {overdeps: {}}
    services:
      ms:
        environment:
          OVERRIDE: oh, the power!
      work:
        volumes: [overdeps:/var/deps]
`

		exp := readFile(t, "../testdata/expectations/registry-user-override.yml")

		assertComposed(t, config, exp, "user preference overrides orders")
	})

	t.Run("config errors", func(t *testing.T) {
		assertComposed(t,
			(preferRegistry + serviceFiles + `
user:
  services:
    microservice: {}
`),
			expRegistry,
			"No error on empty service config")

		assertConfigError(t,
			`
service_files:
  - ../testdata/microservice.yml
user:
  services:
    microservice:
      config: not-found
`,
			"Config 'not-found' for service 'microservice' does not exist",
			"Errors for not-found user service config choice")
	})
}
