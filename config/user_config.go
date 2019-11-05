package config

import (
	"fmt"
)

type UserServiceConfig struct {
	Config   string `mapstructure:"config"`
	Disabled bool   `mapstructure:"disabled"`
}

// UserConfig represents the user's customization file.
type UserConfig struct {
	ServicePreference []string                     `mapstructure:"service_preference"`
	Services          map[string]UserServiceConfig `mapstructure:"services"`
	Override          map[string]interface{}       `mapstructure:"override"`
}

func NewUserConfig() *UserConfig {
	return &UserConfig{}
}

func (cfg *UserConfig) ToMap() (map[string]interface{}, error) {
	cfgMap, err := structToMap(cfg)
	if err != nil {
		return nil, fmt.Errorf("error loading user config: %s", err)
	}
	return cfgMap, nil
}

func UserConfigFromMap(cfgMap map[string]interface{}) (*UserConfig, error) {
	result := NewUserConfig()
	if err := mapToStruct(cfgMap, &result); err != nil {
		return nil, err
	}
	return result, nil
}
