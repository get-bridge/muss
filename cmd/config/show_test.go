package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"gerrit.instructure.com/muss/config"
)

func testShowCommand(t *testing.T, args []string) (string, string) {
	var stdout, stderr strings.Builder

	cmd := newShowCommand()
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs(args)

	if err := cmd.Execute(); err != nil {
		t.Fatal("error executing cmd: ", err)
	}

	return stdout.String(), stderr.String()
}

func showOut(t *testing.T, format string) string {
	stdout, stderr := testShowCommand(t, []string{"--format", format})

	if stderr != "" {
		t.Fatal("error processing template: ", stderr)
	}

	return stdout
}

func showErr(t *testing.T, format string) string {
	stdout, stderr := testShowCommand(t, []string{"--format", format})

	if stdout != "" {
		t.Fatal("stdout:", stdout)
	}

	return stderr
}

func TestConfigShow(t *testing.T) {

	exp := map[string]interface{}{
		"version": "3.5",
		"volumes": map[string]interface{}{},
		"services": map[string]interface{}{
			"app": map[string]interface{}{
				"image": "alpine",
				"environment": map[string]interface{}{
					"FOO": "bar",
				},
				"volumes": []string{
					"./here:/there",
				},
			},
			"store": map[string]interface{}{
				"volumes": []string{
					"data:/var/data",
				},
			},
		},
	}
	cfg := map[string]interface{}{
		"user": map[string]interface{}{
			"services": map[string]interface{}{
				"foo": map[string]interface{}{
					"config": "bar",
				},
			},
			"service_preference": []string{"repo", "registry"},
		},
		"service_definitions": []config.ServiceDef{
			map[string]interface{}{
				"name": "app",
				"configs": map[string]interface{}{
					"sole": exp,
				},
			},
		},
	}

	t.Run("config show", func(t *testing.T) {
		config.SetConfig(cfg)

		assert.Equal(t,
			"3.5",
			showOut(t, "{{ compose.version }}"),
			"basic")

		assert.Equal(t,
			"<FOO = bar>",
			showOut(t, `{{ range compose.services }}{{ range $k, $v := .environment }}<{{ $k }} = {{ $v }}>{{ end }}{{ end }}`),
			"iterate over compose services")

		assert.Equal(t,
			"./here:/there\ndata:/var/data\n",
			showOut(t, `{{ range .service_definitions }}{{ range .configs }}{{ range .services }}{{ range .volumes }}{{ . }}{{ "\n" }}{{ end }}{{ end }}{{ end }}{{ end }}`),
			"iterate over project config service_definitions")

		assert.Equal(t,
			"- ./here:/there\n- data:/var/data\n",
			showOut(t, `{{ range .service_definitions }}{{ range .configs }}{{ range .services }}{{ yaml .volumes }}{{ end }}{{ end }}{{ end }}`),
			"yaml template function")

		assert.Equal(t,
			"repo\nregistry\n",
			showOut(t, `{{ range user.service_preference }}{{ . }}{{ "\n" }}{{ end }}`),
			"user func")

		assert.Equal(t,
			"bar",
			showOut(t, `{{ range .user.services }}{{ .config }}{{ end }}`),
			".user (key)")
	})

	t.Run("empty config", func(t *testing.T) {
		config.SetConfig(map[string]interface{}{})

		assert.Equal(t,
			"",
			showOut(t, `{{ range user.services }}{{ .config }}{{ end }}`),
			"empty user")
	})

	t.Run("config show errors", func(t *testing.T) {
		config.SetConfig(map[string]interface{}{
			"user": map[string]interface{}{
				"services": func() {},
			},
		})

		assert.Contains(t,
			showErr(t, `{{`),
			`unexpected unclosed action in command`,
			"error on stderr",
		)

		assert.Contains(t,
			showErr(t, `{{ yaml .user }}`),
			`cannot marshal type: func()`,
			"yaml() function error on stderr",
		)
	})
}
