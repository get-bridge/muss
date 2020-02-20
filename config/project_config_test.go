package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"gerrit.instructure.com/muss/testutil"
)

func newTestConfig(t *testing.T, cfgMap map[string]interface{}) *ProjectConfig {
	cfg, err := NewConfigFromMap(cfgMap)
	if err != nil {
		t.Fatalf("unexpected config error: %s", err)
	}

	// Just to ensure better coverage, ensure that what we bring in can come out.
	_, err = cfg.ToMap()
	if err != nil {
		t.Fatalf("error translating cfg to map: %s", err)
	}

	return cfg
}

func TestProjectToMap(t *testing.T) {
	t.Run("minimal", func(t *testing.T) {
		cfg := newTestConfig(t, map[string]interface{}{})
		m, err := cfg.ToMap()
		assert.Nil(t, err)

		exp := map[string]interface{}{
			"project_name":               "",
			"compose_file":               "",
			"default_service_preference": []interface{}{},
			"secret_commands":            map[string]interface{}{},
			"secret_passphrase":          "",
			"service_definitions":        []interface{}{},
			"service_files":              []interface{}{},
			"status":                     nil,
			"user":                       nil,
			"user_file":                  "muss.user.yaml",
		}

		assert.Equal(t, exp, m)
	})

	t.Run("allthethings", func(t *testing.T) {
		testutil.WithTempDir(t, func(tmpdir string) {
			os.Unsetenv("MUSS_FILE")
			os.Unsetenv("MUSS_USER_FILE")
			testutil.WriteFile(t, "muss.yaml", `
project_name: tester
compose_file: foo.yml
user_file: user.yml
default_service_preference:
- foo
- bar
secret_commands:
  shh:
    exec: [echo, secret]
    env_commands:
    - varname: MUSS_TEST_TOKEN
      exec: [echo, MUSS_TEST_TOKEN=1]
secret_passphrase: $MUSS_TEST_TOKEN
service_files:
- sd.yml
status:
  exec: [date]
  line_format: "# %s"
  interval: 10s
`)
			testutil.WriteFile(t, "user.yml", `
service_preference:
  - baz
override:
  version: '3.2'
  services:
    s1:
      environment:
        HOGE: piyo
services:
  sd:
    config: bar
  s2:
    disabled: true
`)
			testutil.WriteFile(t, "sd.yml", `
name: sd
configs:
  _base:
    version: '3.1'
  foo:
    secrets:
      FOO_TOKEN: {shh: ["token"]}
    include:
      - _base
      - file: sd-base.yml
  bar:
    include:
      - _base
    volumes: {}
`)
			exp := map[string]interface{}{
				"project_name":               "tester",
				"compose_file":               "foo.yml",
				"user_file":                  "user.yml",
				"default_service_preference": []interface{}{"foo", "bar"},
				"secret_commands": map[string]interface{}{
					"shh": map[string]interface{}{
						"exec": []interface{}{"echo", "secret"},
						"env_commands": []interface{}{
							map[string]interface{}{
								"varname": "MUSS_TEST_TOKEN",
								"exec":    []interface{}{"echo", "MUSS_TEST_TOKEN=1"},
							},
						},
					},
				},
				"secret_passphrase": "$MUSS_TEST_TOKEN",
				"service_definitions": []interface{}{
					map[string]interface{}{
						"name": "sd",
						"file": "sd.yml",
						"configs": map[string]interface{}{
							"_base": map[string]interface{}{
								"version": "3.1",
							},
							"foo": map[string]interface{}{
								"include": []interface{}{
									"_base",
									map[string]interface{}{
										"file": "sd-base.yml",
									},
								},
								"secrets": map[string]interface{}{
									"FOO_TOKEN": map[string]interface{}{
										"shh": []interface{}{"token"},
									},
								},
							},
							"bar": map[string]interface{}{
								"include": []interface{}{
									"_base",
								},
								"volumes": map[string]interface{}{},
							},
						},
					},
				},
				"service_files": []interface{}{"sd.yml"},
				"status": map[string]interface{}{
					"exec":        []interface{}{"date"},
					"line_format": "# %s",
					"interval":    "10s",
				},
				"user": map[string]interface{}{
					"service_preference": []interface{}{"baz"},
					"override": map[string]interface{}{
						"version": "3.2",
						"services": map[string]interface{}{
							"s1": map[string]interface{}{
								"environment": map[string]interface{}{
									"HOGE": "piyo",
								},
							},
						},
					},
					"services": map[string]interface{}{
						"sd": map[string]interface{}{
							"config":   "bar",
							"disabled": false,
						},
						"s2": map[string]interface{}{
							"config":   "",
							"disabled": true,
						},
					},
				},
			}
			cfg, err := NewConfigFromDefaultFile()
			assert.Nil(t, err)
			m, err := cfg.ToMap()
			assert.Nil(t, err)

			assert.Equal(t, exp, m, "round trip")
		})
	})
}
