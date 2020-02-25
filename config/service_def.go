package config

import (
	"fmt"
	"os"
)

// ServiceDef represents a service definition read from a file.
type ServiceDef struct {
	Configs map[string]interface{} `yaml:"configs"`
	Name    string                 `yaml:"name"`
}

func newServiceDef() *ServiceDef {
	return &ServiceDef{}
}

func (s *ServiceDef) chooseConfig(cfg *ProjectConfig) (map[string]interface{}, error) {
	options := s.configOptions()
	result := map[string]interface{}{}

	// Check if user configured this service specifically.
	userChoice := ""
	if cfg.User != nil {
		if userserv, ok := cfg.User.Services[s.Name]; ok {
			if userserv.Disabled {
				return result, nil
			}

			userChoice = userserv.Config
			if userChoice != "" {
				if _, ok := s.Configs[userChoice]; !ok {
					return nil, fmt.Errorf("Config '%s' for service '%s' does not exist", userChoice, s.Name)
				}
			}
		}
	}

	if envChoice := os.Getenv("MUSS_SERVICE_PREFERENCE"); envChoice != "" && s.Configs[envChoice] != nil {
		// If specified via env var, use it.
		result = s.Configs[envChoice].(map[string]interface{})
	} else if userChoice != "" {
		// If user chose specifically, use it.
		result = s.Configs[userChoice].(map[string]interface{})
	} else if len(options) == 1 {
		// If there is only one option, use it.
		result = s.Configs[options[0]].(map[string]interface{})
	} else {
		// To determine which config option to use we can build a list...
		// starting with any user configured preference...
		var order []string
		if cfg.User != nil {
			order = cfg.User.ServicePreference
		} else {
			order = []string{}
		}
		// followed by any project defaults...
		order = append(order, cfg.DefaultServicePreference...)

		// then iterate and use the first preference that this service defines.
		for _, o := range order {
			if found, ok := s.Configs[o]; ok {
				result = found.(map[string]interface{})
				break
			}
		}
	}

	// TODO: recurse
	if includes, ok := result["include"].([]interface{}); ok {
		delete(result, "include")
		base := map[string]interface{}{}
		for _, i := range includes {
			base = mapMerge(base, s.Configs[i.(string)].(map[string]interface{}))
		}
		result = mapMerge(base, result)
	}
	return result, nil
}

func (s *ServiceDef) configOptions() []string {
	keys := make([]string, 0, len(s.Configs))
	for k := range s.Configs {
		if k[0:1] != "_" {
			keys = append(keys, k)
		}
	}
	return keys
}
