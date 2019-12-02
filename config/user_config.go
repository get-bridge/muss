package config

import (
	"fmt"
)

// UserServiceConfig represents the user's configuration for a service
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

// NewUserConfig returns new UserConfig
func NewUserConfig() *UserConfig {
	return &UserConfig{}
}

// ToMap returns new map[string]interface{} from UserConfig
func (cfg *UserConfig) ToMap() (map[string]interface{}, error) {
	cfgMap, err := structToMap(cfg)
	if err != nil {
		return nil, fmt.Errorf("error loading user config: %s", err)
	}
	return cfgMap, nil
}

// UserConfigFromMap returns new UserConfig from map[string]interface{}
func UserConfigFromMap(cfgMap map[string]interface{}) (*UserConfig, error) {
	result := NewUserConfig()
	if err := mapToStruct(cfgMap, &result); err != nil {
		return nil, err
	}
	return result, nil
}
