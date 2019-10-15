package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// NOTE: For YAML:
// - don't let any tabs get in
// - file paths are relative to this test file's parent dir

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
	t.Helper()

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

	expRepo := readTestFile(t, "../testdata/expectations/repo.yml")
	expRegistry := readTestFile(t, "../testdata/expectations/registry.yml")

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

		exp := readTestFile(t, "../testdata/expectations/user-registry-ms-repo.yml")

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

		exp := readTestFile(t, "../testdata/expectations/registry-user-override.yml")

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

	t.Run("secrets", func(t *testing.T) {
		// We don't actually create this we just want a string.
		setCacheRoot("/tmp/.muss-test-cache")

		projectSecrets = nil

		os.Setenv("MUSS_TEST_PASSPHRASE", "decomposing")
		config := preferRegistry + serviceFiles + `
secrets:
  passphrase: $MUSS_TEST_PASSPHRASE
  commands:
    print:
      exec: [echo]
user:
  service_preference: [repo]
  services:
    microservice:
      config: remote
`

		exp := readTestFile(t, "../testdata/expectations/user-registry-ms-remote.yml")

		assertComposed(t, config, exp, "service defs with secrets")

		if len(projectSecrets) != 2 {
			t.Fatalf("expected 2 secrets, found %d", len(projectSecrets))
		}
		assert.Equal(t, "MSKEY", projectSecrets[0].VarName())
		assert.Equal(t, "OTHER_SECRET_TEST", projectSecrets[1].VarName())

		assertComposed(t,
			`
secrets:
  passphrase: $MUSS_TEST_PASSPHRASE
service_definitions:
- name: one
  configs:
    sole:
      secrets:
        - varname: FOO_SECRET
          exec: [echo, foo]
- name: two
  configs:
    sole:
      secrets:
        BAR_SHH:
          exec: [echo, bar]
`,
			`{services: {}, volumes: {}, version: '3.7'}`,
			"secrets as map or list",
		)

		if len(projectSecrets) != 2 {
			t.Fatalf("expected 2 secrets, found %d", len(projectSecrets))
		}
		assert.Equal(t, "FOO_SECRET", projectSecrets[0].VarName())
		assert.Equal(t, "BAR_SHH", projectSecrets[1].VarName())
	})
}

func TestPrepareVolumes(t *testing.T) {
	os.Unsetenv("MUSS_TEST_VAR")

	t.Run("using ./", func(t *testing.T) {
		assertPreparedVolumes(
			t,
			map[string]interface{}{
				"volumes": []interface{}{
					"named_vol:/some/vol",
					"named_child:/usr/src/app/named_mount",
					"/root/dir:/some/root",
					"/root/sub:/usr/src/app/sub/root",
					"${MUSS_TEST_VAR:-.}:/usr/src/app", // keep this in the middle (not first)
					map[string]interface{}{
						"type":   "volume",
						"source": "named_map",
						"target": "/named/map",
					},
					map[string]interface{}{
						"type":   "bind",
						"source": "/file",
						"target": "/anywhere",
						"file":   true,
					},
				},
			},
			FileGenMap{
				"named_mount": ensureExistsOrDir,
				"/root/dir":   ensureExistsOrDir,
				"/root/sub":   ensureExistsOrDir,
				"sub/root":    ensureExistsOrDir,
				"/file":       ensureFile,
			},
		)
	})

	t.Run("using children of ./", func(t *testing.T) {
		assertPreparedVolumes(
			t,
			map[string]interface{}{
				"volumes": []interface{}{
					"${MUSS_TEST_VAR:-./foo}:/somewhere/foo",
					"./bar/:/usr/src/app/bar",
					"named_foo:/somewhere/foo/baz",
					"./t/qux:/usr/src/app/bar/quxt",
				},
			},
			FileGenMap{
				"foo":      ensureExistsOrDir,
				"bar":      ensureExistsOrDir,
				"foo/baz":  ensureExistsOrDir,
				"t/qux":    ensureExistsOrDir,
				"bar/quxt": ensureExistsOrDir,
			},
		)
	})
}

func assertPreparedVolumes(t *testing.T, service map[string]interface{}, exp FileGenMap) {
	t.Helper()

	actual, err := prepareVolumes(service)
	if err != nil {
		t.Fatal(err)
	}

	// Equality assertion doesn't work on func refs so do it another way.
	assert.Equal(t, len(exp), len(actual))
	for k := range exp {
		assert.NotNilf(t, actual[k], "actual %s not nil", k)
		assert.Equalf(t, describeFunc(exp[k]), describeFunc(actual[k]), "funcs for %s", k)
	}

	// However, this diff can be useful when debugging.
	if t.Failed() {
		assert.Equal(t, exp, actual)
	}
}

// Print the function address: "(func(string) error)(0x12de840)"
func describeFunc(v interface{}) string {
	return fmt.Sprintf("%#v", v)
}
