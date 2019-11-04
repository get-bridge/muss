package config

// UserConfig represents the user's customization file.
type UserConfig struct {
	ServicePreference []string `mapstructure:"service_preference"`
	Services          map[string]struct {
		Config   string
		Disabled bool
	}
	Override map[string]interface{}
}
