package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"gerrit.instructure.com/muss/testutil"
)

func TestConfigLoad(t *testing.T) {
	testutil.WithTempDir(t, func(tmpdir string) {

		testutil.WriteFile(t, defaultProjectFile, `project_name: stinky`)
		testutil.WriteFile(t, defaultUserFile, `override: {version: "1.1"}`)

		customProjectFile := "mussy.yml"
		customUserFile := "user.yml"
		testutil.WriteFile(t, customProjectFile, `project_name: dexter`)
		testutil.WriteFile(t, customUserFile, `override: {version: "1.2"}`)

		projectFileWithUser := "musswithuser.yaml"
		userFile2 := "user2.yml"
		testutil.WriteFile(t, projectFileWithUser, `user_file: user.yml`)
		testutil.WriteFile(t, userFile2, `override: {version: "1.3"}`)

		t.Run("no env", func(t *testing.T) {
			os.Unsetenv("MUSS_FILE")
			os.Unsetenv("MUSS_USER_FILE")

			cfg, err := NewConfigFromDefaultFile()

			assert.Nil(t, err, "no error")

			assert.Equal(t, "muss.yaml", cfg.ProjectFile, "default project file")
			assert.Equal(t, cfg.ProjectName, "stinky")

			assert.Equal(t, "muss.user.yaml", cfg.UserFile, "default user file")
			assert.Equal(t, cfg.User.Override["version"].(string), "1.1")
		})

		t.Run("env vars", func(t *testing.T) {
			os.Setenv("MUSS_FILE", customProjectFile)
			os.Setenv("MUSS_USER_FILE", customUserFile)

			cfg, err := NewConfigFromDefaultFile()

			assert.Nil(t, err, "no error")

			assert.Equal(t, customProjectFile, cfg.ProjectFile, "env project file")
			assert.Equal(t, "dexter", cfg.ProjectName)

			assert.Equal(t, customUserFile, cfg.UserFile, "env user file")
			assert.Equal(t, "1.2", cfg.User.Override["version"].(string))
		})

		t.Run("project config sets user file", func(t *testing.T) {
			os.Setenv("MUSS_FILE", projectFileWithUser)
			os.Unsetenv("MUSS_USER_FILE")

			cfg, err := NewConfigFromDefaultFile()

			assert.Nil(t, err, "no error")

			assert.Equal(t, projectFileWithUser, cfg.ProjectFile, "env project file")
			assert.Equal(t, "user.yml", cfg.UserFile, "project-set user file")
			assert.Equal(t, "1.2", cfg.User.Override["version"].(string))
		})

		t.Run("env overrides project", func(t *testing.T) {
			os.Setenv("MUSS_FILE", projectFileWithUser)
			os.Setenv("MUSS_USER_FILE", userFile2)

			cfg, err := NewConfigFromDefaultFile()

			assert.Nil(t, err, "no error")

			assert.Equal(t, projectFileWithUser, cfg.ProjectFile, "env project file")
			assert.Equal(t, userFile2, cfg.UserFile, "env overrides project")
			assert.Equal(t, "1.3", cfg.User.Override["version"].(string))
		})
	})
}
