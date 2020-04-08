package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigWarnings(t *testing.T) {
	t.Run("Warn()", func(t *testing.T) {
		cfg := newTestConfig(t, map[string]interface{}{})

		assert.Nil(t, cfg.Warnings)

		cfg.Warn("foo")

		assert.Equal(t, []string{"foo"}, cfg.Warnings)

		cfg.Warn("bar\nbaz")

		assert.Equal(t, []string{"foo", "bar\nbaz"}, cfg.Warnings)

		cfg.Warn("foo")

		assert.Equal(t, []string{"foo", "bar\nbaz"}, cfg.Warnings, "no duplicates")
	})
}
