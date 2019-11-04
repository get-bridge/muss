package config

// DefaultFile is the name of the default config file.
var DefaultFile = "muss.yaml"

// ProjectFile is the path to the project config file.
var ProjectFile = DefaultFile

// UserFile is the path to the user config file, if defined.
var UserFile string

var project *ProjectConfig

// All returns the whole project config.
func All() *ProjectConfig {
	if project == nil {
		load()
	}
	return project
}
