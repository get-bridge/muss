package config

import (
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	rootcmd "gerrit.instructure.com/muss/cmd"
	"gerrit.instructure.com/muss/config"
)

func testShowCommand(t *testing.T, cfg *config.ProjectConfig, args []string) (string, string) {
	t.Helper()

	var stdout, stderr strings.Builder

	cmd := rootcmd.NewRootCommand(cfg)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	rootcmd.ExecuteRoot(cmd, append([]string{"config", "show"}, args...))

	return stdout.String(), stderr.String()
}

func showOut(t *testing.T, cfg *config.ProjectConfig, format string) string {
	t.Helper()

	stdout, stderr := testShowCommand(t, cfg, []string{"--format", format})

	if stderr != "" {
		t.Fatal("error processing template: ", stderr)
	}

	return stdout
}

func showErr(t *testing.T, cfg *config.ProjectConfig, format string) string {
	t.Helper()

	stdout, stderr := testShowCommand(t, cfg, []string{"--format", format})

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
	cfgMap := map[string]interface{}{
		"user": map[string]interface{}{
			"services": map[string]interface{}{
				"foo": map[string]interface{}{
					"config": "bar",
				},
			},
			"service_preference": []string{"repo", "registry"},
		},
		"service_definitions": []config.ServiceDef{
			config.ServiceDef{
				"name": "app",
				"configs": map[string]interface{}{
					"sole": exp,
				},
			},
		},
	}

	t.Run("config show", func(t *testing.T) {
		config.SetConfig(cfgMap)
		cfg, _ := config.All()

		assert.Equal(t,
			"3.5",
			showOut(t, cfg, "{{ compose.version }}"),
			"basic")

		assert.Equal(t,
			"<FOO = bar>",
			showOut(t, cfg, `{{ range compose.services }}{{ range $k, $v := .environment }}<{{ $k }} = {{ $v }}>{{ end }}{{ end }}`),
			"iterate over compose services")

		assert.Equal(t,
			"./here:/there\ndata:/var/data\n",
			showOut(t, cfg, `{{ range .service_definitions }}{{ range .configs }}{{ range .services }}{{ range .volumes }}{{ . }}{{ "\n" }}{{ end }}{{ end }}{{ end }}{{ end }}`),
			"iterate over project config service_definitions")

		assert.Equal(t,
			"- ./here:/there\n- data:/var/data\n",
			showOut(t, cfg, `{{ range .service_definitions }}{{ range .configs }}{{ range .services }}{{ yaml .volumes }}{{ end }}{{ end }}{{ end }}`),
			"yaml template function")

		assert.Equal(t,
			"repo\nregistry\n",
			showOut(t, cfg, `{{ range user.service_preference }}{{ . }}{{ "\n" }}{{ end }}`),
			"user func")

		assert.Equal(t,
			"bar",
			showOut(t, cfg, `{{ range .user.services }}{{ .config }}{{ end }}`),
			".user (key)")
	})

	t.Run("without service defs", func(t *testing.T) {
		if dir, err := os.Getwd(); err != nil {
			t.Fatal(err)
		} else {
			defer os.Chdir(dir)
			os.Chdir(path.Join(dir, "..", "..", "testdata", "no-muss"))
		}

		config.SetConfig(nil)
		cfg, _ := config.All()

		stdout, stderr := testShowCommand(t, cfg, []string{"--format", `{{ range $k, $v := compose.services }}{{ $k }}{{ "." }}{{ $v.image }}{{ "\n" }}{{ end }}`})

		assert.Equal(t,
			"a1.alpine\na2.alpine\n",
			stdout,
			"compose config without service defs")

		assert.Equal(t,
			"muss project config 'muss.yaml' file not found.\n",
			stderr,
			"warns about no project config")
	})

	t.Run("empty config", func(t *testing.T) {
		config.SetConfig(map[string]interface{}{})
		cfg, _ := config.All()

		assert.Equal(t,
			"",
			showOut(t, cfg, `{{ range user.services }}{{ .config }}{{ end }}`),
			"empty user")
	})

	t.Run("config show errors", func(t *testing.T) {
		config.SetConfig(map[string]interface{}{
			"user": map[string]interface{}{
				"override": map[string]interface{}{
					"yamlbreaker": func() {},
				},
			},
		})
		cfg, _ := config.All()

		assert.Contains(t,
			showErr(t, cfg, `{{`),
			`unexpected unclosed action in command`,
			"error on stderr",
		)

		assert.Contains(t,
			showErr(t, cfg, `{{ yaml .user }}`),
			`cannot marshal type: func()`,
			"yaml() function error on stderr",
		)
	})
}
