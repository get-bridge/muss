package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigEnv(t *testing.T) {
	t.Run("no config", func(t *testing.T) {
		SetConfig(map[string]interface{}{})

		cfg, err := All()
		if err != nil {
			t.Fatal(err)
		}

		assert.Nil(t, cfg.LoadEnv(), "no errors")
	})

	t.Run("project_name", func(t *testing.T) {
		os.Unsetenv("COMPOSE_PROJECT_NAME")
		defer os.Unsetenv("COMPOSE_PROJECT_NAME")

		SetConfig(map[string]interface{}{
			"project_name": "musstest",
		})

		cfg, err := All()
		if err != nil {
			t.Fatal(err)
		}

		assert.Nil(t, cfg.LoadEnv(), "no errors")
		assert.Equal(t, os.Getenv("COMPOSE_PROJECT_NAME"), "musstest")

		cfg.ProjectName = "nerts"

		assert.Nil(t, cfg.LoadEnv(), "no errors")
		assert.Equal(t, os.Getenv("COMPOSE_PROJECT_NAME"), "musstest", "doesn't overwrite")

		os.Unsetenv("COMPOSE_PROJECT_NAME")
		assert.Nil(t, cfg.LoadEnv(), "no errors")
		assert.Equal(t, os.Getenv("COMPOSE_PROJECT_NAME"), "nerts", "sets when not set")
	})

	t.Run("secrets", func(t *testing.T) {
		os.Unsetenv("MUSS_TEST_ENV")
		defer os.Unsetenv("MUSS_TEST_ENV")

		SetConfig(map[string]interface{}{})

		cfg, err := All()
		if err != nil {
			t.Fatal(err)
		}
		cfg.Secrets = append(cfg.Secrets, &envCmd{
			parse: true,
			exec:  []string{"/bin/sh", "-c", "echo MUSS_TEST_ENV=42"},
		})

		assert.Nil(t, cfg.LoadEnv(), "no errors")
		assert.Equal(t, os.Getenv("MUSS_TEST_ENV"), "42")
	})

	t.Run("returns error", func(t *testing.T) {
		SetConfig(map[string]interface{}{})

		cfg, err := All()
		if err != nil {
			t.Fatal(err)
		}
		cfg.Secrets = append(cfg.Secrets, &envCmd{
			parse: true,
			exec:  []string{"false"},
		})

		loadErr := cfg.LoadEnv()
		if loadErr == nil {
			t.Fatal("expected error, got nil")
		}
		assert.Equal(t, loadErr.Error(), "Failed to load secrets: command failed: exit status 1")
	})
}
