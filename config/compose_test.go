package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"gerrit.instructure.com/muss/testutil"
)

// NOTE: For YAML:
// - don't let any tabs get in
// - file paths are relative to this test file's parent dir

func parseAndCompose(yaml string) (map[string]interface{}, *ProjectConfig, error) {
	parsed, err := parseYaml([]byte(yaml))
	if err != nil {
		return nil, nil, err
	}

	cfg, err := NewConfigFromMap(parsed)
	if err != nil {
		return nil, nil, err
	}

	dc, err := cfg.ComposeConfig()
	if err != nil {
		return nil, nil, err
	}

	// Just for better test coverage... ensure anything we put in can come out.
	_, err = cfg.ToMap()
	if err != nil {
		return nil, nil, err
	}

	return dc, cfg, nil
}

func assertConfigError(t *testing.T, config, expErr string, msgAndArgs ...interface{}) {
	t.Helper()

	_, _, err := parseAndCompose(config)
	if err == nil {
		t.Fatal("expected error, found nil")
	}
	assert.Contains(t, err.Error(), expErr, msgAndArgs...)
}

func assertComposed(t *testing.T, config, exp string, msgAndArgs ...interface{}) *ProjectConfig {
	t.Helper()

	parsedExp, err := parseYaml([]byte(exp))
	if err != nil {
		t.Fatalf("Error parsing exp yaml: %s", err)
	}

	actual, projectConfig, err := parseAndCompose(config)
	if err != nil {
		t.Fatalf("Error parsing config: %s", err)
	}

	assert.Equal(t,
		parsedExp,
		actual,
		msgAndArgs...)

	return projectConfig
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

	secretConfig := `
secret_passphrase: $MUSS_TEST_PASSPHRASE
secret_commands:
  print:
    exec: [echo]
  show:
    exec: [echo]
`

	expRepo := testutil.ReadFile(t, "../testdata/expectations/repo.yml")
	expRegistry := testutil.ReadFile(t, "../testdata/expectations/registry.yml")
	expRepoMsRemote := testutil.ReadFile(t, "../testdata/expectations/user-repo-ms-remote.yml")
	expRegistryMsRepo := testutil.ReadFile(t, "../testdata/expectations/user-registry-ms-repo.yml")

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

		assertComposed(t, config, expRegistryMsRepo, "user preference overrides orders")
	})

	t.Run("env var custom service config", func(t *testing.T) {
		config := preferRepo + serviceFiles
		defer os.Unsetenv("MUSS_SERVICE_PREFERENCE")

		os.Setenv("MUSS_SERVICE_PREFERENCE", "registry")
		assertComposed(t, config, expRegistry, "MUSS_SERVICE_PREFERENCE env var overrides orders")
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

		exp := testutil.ReadFile(t, "../testdata/expectations/registry-user-override.yml")

		assertComposed(t, config, exp, "user preference overrides orders")
	})

	t.Run("user can disable services", func(t *testing.T) {
		config := preferRegistry + serviceFiles + `
user:
  services:
    app:
      disabled: true
`

		exp := testutil.ReadFile(t, "../testdata/expectations/user-registry-app-disabled.yml")

		assertComposed(t, config, exp, "user disabled services")
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

		os.Setenv("MUSS_TEST_PASSPHRASE", "decomposing")
		config := preferRegistry + serviceFiles + secretConfig + `
user:
  service_preference: [repo]
  services:
    microservice:
      config: remote
`

		projectConfig := assertComposed(t, config, expRepoMsRemote, "service defs with secrets")

		if len(projectConfig.Secrets) != 3 {
			t.Fatalf("expected 3 secrets, found %d", len(projectConfig.Secrets))
		}
		assert.Equal(t, "MSKEY", projectConfig.Secrets[0].VarName())
		assert.Equal(t, "OTHER_SECRET_TEST", projectConfig.Secrets[1].VarName())

		projectConfig = assertComposed(t,
			`
secret_passphrase: $MUSS_TEST_PASSPHRASE
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
        SECOND_BAR:
          exec: [echo, two]
`,
			`{version: '3.7'}`,
			"secrets as map or list",
		)

		actualVarNames := make([]string, len(projectConfig.Secrets))
		for i := range projectConfig.Secrets {
			actualVarNames[i] = projectConfig.Secrets[i].VarName()
		}

		assert.ElementsMatch(t,
			[]string{"FOO_SECRET", "BAR_SHH", "SECOND_BAR"},
			actualVarNames)
	})

	t.Run("include errors", func(t *testing.T) {
		assertConfigError(t, `
service_definitions:
- name: one
  configs:
    _base:
      version: '2.1'
    sole:
      include:
        - _no
`,
			"invalid 'include'; config '_no' not found",
			"bad include string")

		assertConfigError(t, `
service_definitions:
- name: one
  configs:
    _base:
      version: '2.1'
    sole:
      include:
        - bad: map
`,
			"invalid 'include' map; valid keys: 'file'",
			"bad include map")

		assertConfigError(t, `
service_definitions:
- name: one
  configs:
    _base:
      version: '2.1'
    sole:
      include:
        - [no, good]
`,
			"invalid 'include' value; must be a string or a map",
			"bad include type")

		assertConfigError(t, `
service_definitions:
- name: one
  configs:
    _base:
      version: '2.1'
    sole:
      include:
        - file: no-file.txt
`,
			"failed to read 'no-file.txt': open no-file.txt: no such file",
			"bad include type")
	})

	t.Run("include", func(t *testing.T) {
		assertComposed(t, `
service_definitions:
- name: one
  configs:
    _base:
      version: '2.1'
    sole:
      include:
        - _base
`,
			"{version: '2.1'}",
			"include string")

		assertComposed(t, `
service_definitions:
- name: one
  configs:
    _base:
      services:
        app:
          image: alpine
          init: true
    _edge:
      services:
        app:
          image: alpine:edge
          tty: true
    sole:
      include:
        - _base
        - _edge
`,
			"{version: '3.7', services: {app: {image: alpine:edge, init: true, tty: true}}}",
			"multiple include strings merge")

		testutil.WithTempDir(t, func(tmpdir string) {
			testutil.WriteFile(t, filepath.Join("files", "between.yml"), `
version: '2.3'
services:
  app:
    image: alpine:latest
    stdin_open: true
`)

			assertComposed(t, `
service_definitions:
- name: one
  file: `+filepath.Join("files", "sd.yml")+`
  configs:
    sole:
      include:
        - file: between.yml
`,
				"{version: '2.3', services: {app: {image: alpine:latest, stdin_open: true}}}",
				"include file")

			assertComposed(t, `
service_definitions:
- name: one
  file: `+filepath.Join("files", "sd.yml")+`
  configs:
    _base:
      services:
        app:
          image: alpine
          init: true
    _edge:
      services:
        app:
          image: alpine:edge
          tty: true
    sole:
      include:
        - _base
        - file: between.yml
        - _edge
`,
				"{version: '2.3', services: {app: {image: alpine:edge, init: true, tty: true, stdin_open: true}}}",
				"include strings and file mixed")
		})

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
