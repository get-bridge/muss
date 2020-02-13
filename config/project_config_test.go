package config

import (
	"testing"
)

func newTestConfig(t *testing.T, cfgMap map[string]interface{}) *ProjectConfig {
	cfg, err := NewConfigFromMap(cfgMap)
	if err != nil {
		t.Fatalf("unexpected config error: %s", err)
	}
	return cfg
}
