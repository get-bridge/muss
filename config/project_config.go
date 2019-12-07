package config

import (
	"fmt"
)

// ProjectConfig is a type for the parsed contents of the project config file.
type ProjectConfig struct {
	ServiceDefinitions       []ServiceDef           `mapstructure:"service_definitions"`
	UserFile                 string                 `mapstructure:"user_file"`
	User                     *UserConfig            `mapstructure:"user"`
	ServiceFiles             []string               `mapstructure:"service_files"`
	SecretCommands           map[string]interface{} `mapstructure:"secret_commands"`
	SecretPassphrase         string                 `mapstructure:"secret_passphrase"`
	DefaultServicePreference []string               `mapstructure:"default_service_preference"`
	Status                   *StatusConfig          `mapstructure:"status"`

	Secrets []envLoader

	composeConfig   map[string]interface{}
	filesToGenerate FileGenMap
}

// NewProjectConfig returns new ProjectConfig
func NewProjectConfig() *ProjectConfig {
	return &ProjectConfig{}
}

// NewProjectConfigFromMap returns new ProjectConfig from map[string]interface{}
func NewProjectConfigFromMap(cfgMap map[string]interface{}) (*ProjectConfig, error) {
	result := NewProjectConfig()
	if err := mapToStruct(cfgMap, result); err != nil {
		return nil, err
	}
	return result, nil
}

// ToMap returns new map[string]interface{} from ProjectConfig
func (cfg *ProjectConfig) ToMap() (map[string]interface{}, error) {
	cfgMap, err := structToMap(cfg)
	if err != nil {
		return nil, fmt.Errorf("error converting project config to map: %s", err)
	}
	return cfgMap, nil
}
