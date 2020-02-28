package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func envIsUnset(key string) bool {
	_, ok := os.LookupEnv(key)
	return !ok
}

func TestConfigEnv(t *testing.T) {
	t.Run("no config", func(t *testing.T) {
		os.Unsetenv("COMPOSE_PROJECT_NAME")
		os.Unsetenv("COMPOSE_FILE")
		cfg := newTestConfig(t, nil)

		assert.Nil(t, cfg.LoadEnv(), "no errors")
		assert.True(t, envIsUnset("COMPOSE_PROJECT_NAME"), "COMPOSE_PROJECT_NAME not set")
		assert.True(t, envIsUnset("COMPOSE_FILE"), "COMPOSE_FILE not set")
	})

	t.Run("project_name", func(t *testing.T) {
		os.Unsetenv("COMPOSE_PROJECT_NAME")
		defer os.Unsetenv("COMPOSE_PROJECT_NAME")

		cfg := newTestConfig(t, map[string]interface{}{
			"project_name": "musstest",
		})

		assert.Nil(t, cfg.LoadEnv(), "no errors")
		assert.Equal(t, os.Getenv("COMPOSE_PROJECT_NAME"), "musstest")

		cfg.ProjectName = "nerts"

		assert.Nil(t, cfg.LoadEnv(), "no errors")
		assert.Equal(t, os.Getenv("COMPOSE_PROJECT_NAME"), "musstest", "doesn't overwrite")

		os.Unsetenv("COMPOSE_PROJECT_NAME")
		assert.Nil(t, cfg.LoadEnv(), "no errors")
		assert.Equal(t, os.Getenv("COMPOSE_PROJECT_NAME"), "nerts", "sets when not set")
	})

	t.Run("compose_file", func(t *testing.T) {
		os.Unsetenv("COMPOSE_FILE")
		defer os.Unsetenv("COMPOSE_FILE")

		cfg := newTestConfig(t, map[string]interface{}{
			"compose_file": "dc.muss.yml",
		})

		assert.Nil(t, cfg.LoadEnv(), "no errors")
		assert.Equal(t, os.Getenv("COMPOSE_FILE"), "dc.muss.yml")

		cfg.ComposeFile = "nerts"

		assert.Nil(t, cfg.LoadEnv(), "no errors")
		assert.Equal(t, os.Getenv("COMPOSE_FILE"), "dc.muss.yml", "doesn't overwrite")

		os.Unsetenv("COMPOSE_FILE")
		assert.Nil(t, cfg.LoadEnv(), "no errors")
		assert.Equal(t, os.Getenv("COMPOSE_FILE"), "nerts", "sets when not set")
	})

	t.Run("secrets", func(t *testing.T) {
		os.Unsetenv("MUSS_TEST_ENV")
		defer os.Unsetenv("MUSS_TEST_ENV")

		cfg := newTestConfig(t, nil)
		cfg.Secrets = append(cfg.Secrets, &EnvCommand{
			Parse: true,
			Exec:  []string{"/bin/sh", "-c", "echo MUSS_TEST_ENV=42"},
		})

		assert.Nil(t, cfg.LoadEnv(), "no errors")
		assert.Equal(t, os.Getenv("MUSS_TEST_ENV"), "42")
	})

	t.Run("returns error", func(t *testing.T) {
		cfg := newTestConfig(t, nil)

		cfg.Secrets = append(cfg.Secrets, &EnvCommand{
			Parse: true,
			Exec:  []string{"false"},
		})

		loadErr := cfg.LoadEnv()
		if loadErr == nil {
			t.Fatal("expected error, got nil")
		}
		assert.Equal(t, loadErr.Error(), "Failed to load secrets: command failed: exit status 1")
	})
}
