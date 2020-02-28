package config

import (
	"fmt"
)

// ProjectConfig is a type for the parsed contents of the project config file.
type ProjectConfig struct {
	ServiceDefinitions       []*ServiceDef             `yaml:"service_definitions"`
	UserFile                 string                    `yaml:"user_file"`
	User                     *UserConfig               `yaml:"user"`
	ServiceFiles             []string                  `yaml:"service_files"`
	SecretCommands           map[string]*SecretCommand `yaml:"secret_commands"`
	SecretPassphrase         string                    `yaml:"secret_passphrase"`
	DefaultServicePreference []string                  `yaml:"default_service_preference"`
	Status                   *StatusConfig             `yaml:"status"`
	ProjectName              string                    `yaml:"project_name"`
	ComposeFile              string                    `yaml:"compose_file"`

	Secrets     []envLoader `yaml:"-"`
	ProjectFile string      `yaml:"-"`
	LoadError   error       `yaml:"-"`
	Warnings    []string    `yaml:"-"`

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
