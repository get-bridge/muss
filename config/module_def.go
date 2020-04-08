package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ModuleDef represents a module definition read from a file.
type ModuleDef struct {
	Configs map[string]interface{} `yaml:"configs"`
	File    string                 `yaml:"file"`
	Name    string                 `yaml:"name"`
}

func newModuleDef(file string) *ModuleDef {
	return &ModuleDef{
		File: file,
	}
}

func (s *ModuleDef) chooseConfig(cfg *ProjectConfig) (map[string]interface{}, error) {
	options := s.configOptions()
	var result map[string]interface{}

	// Check if user configured this module specifically.
	userChoice := ""
	if cfg.User != nil {
		if userserv, ok := cfg.User.Modules[s.Name]; ok {
			if userserv.Disabled {
				return map[string]interface{}{}, nil
			}

			userChoice = userserv.Config
			if userChoice != "" {
				if _, ok := s.Configs[userChoice]; !ok {
					return nil, fmt.Errorf("Config '%s' for module '%s' does not exist", userChoice, s.Name)
				}
			}
		}
	}

	order := make([]string, 0)
	if envChoice := os.Getenv("MUSS_MODULE_ORDER"); envChoice != "" {
		// If specified via env var, use it.
		order = append(order, strings.Split(envChoice, ",")...)
	} else if envChoice := os.Getenv("MUSS_SERVICE_PREFERENCE"); envChoice != "" {
		cfg.Warn("MUSS_SERVICE_PREFERENCE is deprecated in favor of MUSS_MODULE_ORDER.")
		order = append(order, envChoice)
	} else if userChoice != "" {
		// If user chose specifically, use it.
		result = s.Configs[userChoice].(map[string]interface{})
	}

	// If there is only one option, use it.
	if len(options) == 1 {
		result = s.Configs[options[0]].(map[string]interface{})
	}

	if result == nil {
		// To determine which config option to use we can build a list...
		// starting with any user configured preference...
		if cfg.User != nil {
			order = append(order, cfg.User.ModuleOrder...)
		}
		// followed by any project defaults...
		order = append(order, cfg.DefaultModuleOrder...)

		// then iterate and use the first preference that this module defines.
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
			var input map[string]interface{}
			if msi, ok := i.(map[string]interface{}); ok {
				if file, ok := msi["file"].(string); ok && file != "" {
					file = filepath.Join(filepath.Dir(s.File), file)
					value, err := readCachedYamlFile(file)
					if err != nil {
						return nil, fmt.Errorf("failed to read '%s': %w", file, err)
					}
					input = value
				} else {
					return nil, errors.New("invalid 'include' map; valid keys: 'file'")
				}
			} else if str, ok := i.(string); ok {
				if value, ok := s.Configs[str].(map[string]interface{}); ok {
					input = value
				} else {
					return nil, fmt.Errorf("invalid 'include'; config '%s' not found", str)
				}
			} else {
				return nil, errors.New("invalid 'include' value; must be a string or a map")
			}
			base = mapMerge(base, input)
		}
		result = mapMerge(base, result)
	}
	return result, nil
}

func (s *ModuleDef) configOptions() []string {
	keys := make([]string, 0, len(s.Configs))
	for k := range s.Configs {
		if k[0:1] != "_" {
			keys = append(keys, k)
		}
	}
	return keys
}
