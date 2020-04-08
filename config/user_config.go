package config

import (
	"fmt"
)

// UserModuleConfig represents the user's configuration for a module.
type UserModuleConfig struct {
	Config   string `yaml:"config"`
	Disabled bool   `yaml:"disabled"`
}

// UserConfig represents the user's customization file.
type UserConfig struct {
	ModuleOrder []string                    `yaml:"module_order"`
	Modules     map[string]UserModuleConfig `yaml:"modules"`
	Override    map[string]interface{}      `yaml:"override"`

	DeprecatedServicePreference []string                    `yaml:"service_preference,omitempty"`
	DeprecatedServices          map[string]UserModuleConfig `yaml:"services,omitempty"`
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
