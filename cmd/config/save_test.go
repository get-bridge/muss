package config

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	rootcmd "github.com/instructure-bridge/muss/cmd"
	"github.com/instructure-bridge/muss/config"
	"github.com/instructure-bridge/muss/testutil"
)

func runConfigSave(cfg *config.ProjectConfig) (string, string, int) {
	cmd := rootcmd.NewRootCommand(cfg)
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	ec := rootcmd.ExecuteRoot(cmd, []string{"config", "save"})
	return stdout.String(), stderr.String(), ec
}

func TestConfigSaveCommand(t *testing.T) {
	t.Run("help description", func(t *testing.T) {
		assert.Equal(t,
			"Generate new docker-compose.yml file.",
			newSaveCommand(nil).Long,
			"default")

		cfg, err := config.NewConfigFromMap(map[string]interface{}{"compose_file": "dc.muss.yml"})
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t,
			"Generate new dc.muss.yml file.",
			newSaveCommand(cfg).Long,
			"default")
	})

	t.Run("config save", func(t *testing.T) {
		testutil.WithTempDir(t, func(dir string) {

			cfg, err := config.NewConfigFromMap(map[string]interface{}{
				"project_name": "s1",
				"module_definitions": []map[string]interface{}{
					map[string]interface{}{
						"name": "app",
						"configs": map[string]interface{}{
							"sole": map[string]interface{}{
								"version": "2.1",
							},
						},
					},
				},
			})
			if err != nil {
				t.Fatal(err)
			}

			path := "docker-compose.yml"

			testutil.NoFileExists(t, path)

			var stdout, stderr string
			var exitCode int

			stdout, stderr, exitCode = runConfigSave(cfg)
			assert.Equal(t, 0, exitCode, "exit 0")
			assert.Equal(t, "", stdout, "no out")
			assert.Equal(t, "", stderr, "no err")
			assert.Contains(t, testutil.ReadFile(t, path), `version: "2.1"`, "config written")

			os.Remove(path)
			testutil.NoFileExists(t, path)
			if err := os.Mkdir(path, 0600); err != nil {
				t.Fatalf("failed to make dir: %s", err)
			}

			stdout, stderr, exitCode = runConfigSave(cfg)
			assert.Equal(t, 1, exitCode, "exit 1")
			assert.Equal(t, "", stdout, "no out")
			assert.Equal(t, "Error:  open "+path+": is a directory\n", stderr, "error, no usage string")
			assert.DirExists(t, path, "still a dir")

		})
	})
}
