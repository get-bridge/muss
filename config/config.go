package config

// DefaultFile is the name of the default config file.
var DefaultFile = "muss.yaml"

// ProjectFile is the path to the project config file.
var ProjectFile = DefaultFile

// UserFile is the path to the user config file, if defined.
var UserFile string

// ProjectConfig is a type for the parsed contents of the project config file.
type ProjectConfig map[string]interface{}

// UserConfig represents the user's customization file.
type UserConfig struct {
	ServicePreference []string `mapstructure:"service_preference"`
	Services          map[string]struct {
		Config   string
		Disabled bool
	}
	Override map[string]interface{}
}

// ServiceDef represents a service definition read from a file.
type ServiceDef map[string]interface{}

var project ProjectConfig

// All returns the whole project config.
func All() ProjectConfig {
	if project == nil {
		load()
	}
	return project
}
