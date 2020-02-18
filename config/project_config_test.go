package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func newTestConfig(t *testing.T, cfgMap map[string]interface{}) *ProjectConfig {
	cfg, err := NewConfigFromMap(cfgMap)
	if err != nil {
		t.Fatalf("unexpected config error: %s", err)
	}
	return cfg
}

func TestProjectToMap(t *testing.T) {
	t.Run("config ToMap", func(t *testing.T) {
		cfg := newTestConfig(t, map[string]interface{}{
			"project_name": "map",
		})
		m, err := cfg.ToMap()
		assert.Nil(t, err)

		assert.Equal(t, "map", m["project_name"])
		_, ok := m["user"].(map[string]interface{})
		assert.True(t, ok)
	})
}
