package config

import (
	"fmt"
)

// ProjectConfig is a type for the parsed contents of the project config file.
type ProjectConfig struct {
	ServiceDefinitions       []*ServiceDef          `mapstructure:"service_definitions"`
	UserFile                 string                 `mapstructure:"user_file"`
	User                     *UserConfig            `mapstructure:"user"`
	ServiceFiles             []string               `mapstructure:"service_files"`
	SecretCommands           map[string]interface{} `mapstructure:"secret_commands"`
	SecretPassphrase         string                 `mapstructure:"secret_passphrase"`
	DefaultServicePreference []string               `mapstructure:"default_service_preference"`
	Status                   *StatusConfig          `mapstructure:"status"`
	ProjectName              string                 `mapstructure:"project_name"`
	ComposeFile              string                 `mapstructure:"compose_file"`

	Secrets     []envLoader
	ProjectFile string   `mapstructure:"-"`
	LoadError   error    `mapstructure:"-"`
	Warnings    []string `mapstructure:"-"`

	composeConfig   map[string]interface{}
	filesToGenerate FileGenMap
}

func newProjectConfig() *ProjectConfig {
	return &ProjectConfig{}
}

// NewConfigFromMap returns new ProjectConfig from map[string]interface{}.
func NewConfigFromMap(cfgMap map[string]interface{}) (*ProjectConfig, error) {
	cfg := newProjectConfig()
	cfg.LoadError = cfg.loadMap(cfgMap)
	return cfg, cfg.LoadError
}

// NewConfigFromDefaultFile will attempt to load the config from files.
func NewConfigFromDefaultFile() (*ProjectConfig, error) {
	cfg := newProjectConfig()
	cfg.loadDefaultFile()
	return cfg, cfg.LoadError
}

// ToMap returns new map[string]interface{} from ProjectConfig
func (cfg *ProjectConfig) ToMap() (map[string]interface{}, error) {
	cfgMap, err := structToMap(cfg)
	if err != nil {
		return nil, fmt.Errorf("error converting project config to map: %s", err)
	}
	return cfgMap, nil
}
