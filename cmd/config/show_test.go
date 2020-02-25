package config

import (
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	rootcmd "gerrit.instructure.com/muss/cmd"
	"gerrit.instructure.com/muss/config"
	"gerrit.instructure.com/muss/testutil"
)

func testShowCommand(t *testing.T, cfg *config.ProjectConfig, args []string) (string, string, int) {
	t.Helper()

	var stdout, stderr strings.Builder

	cmd := rootcmd.NewRootCommand(cfg)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)

	exitCode := rootcmd.ExecuteRoot(cmd, append([]string{"config", "show"}, args...))

	return stdout.String(), stderr.String(), exitCode
}

func showOut(t *testing.T, cfg *config.ProjectConfig, format string) string {
	t.Helper()

	stdout, stderr, ec := testShowCommand(t, cfg, []string{"--format", format})

	if stderr != "" {
		t.Fatal("error processing template: ", stderr)
	}

	if ec != 0 {
		t.Fatal("exited non zero: ", ec)
	}

	return stdout
}

func showErr(t *testing.T, cfg *config.ProjectConfig, format string) string {
	t.Helper()

	stdout, stderr, ec := testShowCommand(t, cfg, []string{"--format", format})

	if stdout != "" {
		t.Fatal("stdout:", stdout)
	}

	if ec != 1 {
		t.Fatal("did not exit 1:", ec)
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
		"project_name": "the_hunt",
		"user": map[string]interface{}{
			"services": map[string]interface{}{
				"foo": map[string]interface{}{
					"config": "bar",
				},
			},
			"service_preference": []string{"repo", "registry"},
		},
		"service_definitions": []map[string]interface{}{
			map[string]interface{}{
				"name": "app",
				"configs": map[string]interface{}{
					"sole": exp,
				},
			},
		},
	}

	t.Run("config show", func(t *testing.T) {
		cfg, err := config.NewConfigFromMap(cfgMap)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t,
			"the_hunt",
			showOut(t, cfg, "{{ .project_name }}"),
			"dot context")

		assert.Equal(t,
			"the_hunt",
			showOut(t, cfg, "{{ project.project_name }}"),
			"project function")

		assert.Equal(t,
			"3.5",
			showOut(t, cfg, "{{ compose.version }}"),
			"compose function")

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

		cfg, _ := config.NewConfigFromMap(nil)

		stdout := showOut(t, cfg, `{{ range $k, $v := compose.services }}{{ $k }}{{ "." }}{{ $v.image }}{{ "\n" }}{{ end }}`)

		assert.Equal(t,
			"a1.alpine\na2.alpine\n",
			stdout,
			"compose config without service defs")
	})

	t.Run("empty config", func(t *testing.T) {
		cfg, _ := config.NewConfigFromMap(nil)

		assert.Equal(t,
			"",
			showOut(t, cfg, `{{ range user.services }}{{ .config }}{{ end }}`),
			"empty user")
	})

	t.Run("config show errors", func(t *testing.T) {
		cfg, err := config.NewConfigFromMap(map[string]interface{}{
			"user": map[string]interface{}{
				"override": map[string]interface{}{},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		stderr := showErr(t, cfg, `{{`)

		assert.Contains(t,
			stderr,
			`unexpected unclosed action in command`,
			"error on stderr",
		)

		assert.NotContains(t, stderr, `Usage:`, "no usage")

		cfg.User.Override["yamlbreaker"] = func() {}
		assert.Contains(t,
			showErr(t, cfg, `{{ yaml .user }}`),
			`cannot marshal type: func()`,
			"yaml() function error on stderr",
		)
	})

	t.Run("show config loading error", func(t *testing.T) {
		cfg, err := config.NewConfigFromMap(map[string]interface{}{
			"user_file": []string{"foo"},
		})

		expErr := `expected type 'string', got`
		assert.Contains(t, err.Error(), expErr)

		stdout, stderr, ec := testShowCommand(t, cfg, []string{"--format", `{{ "hi" }}`})

		assert.Contains(t,
			stderr,
			expErr,
			"error on stderr",
		)

		assert.Equal(t, "hi", stdout, "format still works")
		assert.Equal(t, 0, ec, "exit 0")
	})

	t.Run("show file load warnings", func(t *testing.T) {
		testutil.WithTempDir(t, func(tmpdir string) {
			t.Run("missing file", func(t *testing.T) {
				testutil.NoFileExists(t, "muss.yaml")
				cfg, err := config.NewConfigFromDefaultFile()

				expErr := `config file 'muss.yaml' not found`
				assert.Equal(t, expErr, err.Error())

				stdout, stderr, ec := testShowCommand(t, cfg, []string{"--format", `{{ "hi" }}`})

				assert.Equal(t,
					"error loading config: "+expErr+"\n",
					stderr,
					"error on stderr",
				)

				assert.Equal(t, "hi", stdout, "format still works")
				assert.Equal(t, 0, ec, "exit 0")
			})

			t.Run("malformed yaml", func(t *testing.T) {
				testutil.WriteFile(t, "muss.yaml", "---\n{")

				cfg, err := config.NewConfigFromDefaultFile()

				expErr := `Failed to read config file 'muss.yaml': yaml: line 2:`
				assert.Contains(t, err.Error(), expErr)

				stdout, stderr, ec := testShowCommand(t, cfg, []string{"--format", `{{ "hi" }}`})

				assert.Contains(t,
					stderr,
					"error loading config: "+expErr,
					"error on stderr",
				)

				assert.Equal(t, "hi", stdout, "format still works")
				assert.Equal(t, 0, ec, "exit 0")
			})
		})
	})
}
