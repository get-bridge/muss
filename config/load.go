package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/mitchellh/mapstructure"
	yaml "gopkg.in/yaml.v2"
)

var defaultProjectFile = "muss.yaml"
var defaultUserFile = "muss.user.yaml"

func (cfg *ProjectConfig) loadDefaultFile() {
	cfg.ProjectFile = os.Getenv("MUSS_FILE")
	if cfg.ProjectFile == "" {
		cfg.ProjectFile = defaultProjectFile
	}

	// If there is no config file do the best you can
	// (allow muss to wrap docker-compose without a config file).
	if _, err := os.Stat(cfg.ProjectFile); err != nil && os.IsNotExist(err) {
		cfg.LoadError = fmt.Errorf("config file '%s' not found", cfg.ProjectFile)
		return
	}

	object, err := readYamlFile(cfg.ProjectFile)
	if err != nil {
		cfg.LoadError = fmt.Errorf("Failed to read config file '%s': %w", cfg.ProjectFile, err)
		return
	}

	cfg.LoadError = cfg.loadMap(object)
}

func (cfg *ProjectConfig) loadMap(object map[string]interface{}) error {
	if err := mapToStruct(object, cfg); err != nil {
		return err
	}

	// Transform deprecated fields.
	if len(cfg.DeprecatedServiceDefinitions) > 0 {
		cfg.Warn("Configuration 'service_definitions' is deprecated in favor of 'module_definitions'.")
		cfg.ModuleDefinitions = append(cfg.DeprecatedServiceDefinitions, cfg.ModuleDefinitions...)
		cfg.DeprecatedServiceDefinitions = nil
	}
	if len(cfg.DeprecatedServiceFiles) > 0 {
		cfg.Warn("Configuration 'service_files' is deprecated in favor of 'module_files'.")
		cfg.ModuleFiles = append(cfg.DeprecatedServiceFiles, cfg.ModuleFiles...)
		cfg.DeprecatedServiceFiles = nil
	}
	if len(cfg.DeprecatedDefaultServicePreference) > 0 {
		cfg.Warn("Configuration 'default_service_preference' is deprecated in favor of 'default_module_order'.")
		cfg.DefaultModuleOrder = append(cfg.DeprecatedDefaultServicePreference, cfg.DefaultModuleOrder...)
		cfg.DeprecatedDefaultServicePreference = nil
	}

	loaded, err := loadModuleDefs(cfg.ModuleFiles)
	if err != nil {
		return err
	}
	cfg.ModuleDefinitions = append(cfg.ModuleDefinitions, loaded...)

	// Prefer env user file if present.
	if envUserFile := os.Getenv("MUSS_USER_FILE"); envUserFile != "" {
		cfg.UserFile = envUserFile
	}
	// If not set by env or project config use default.
	if cfg.UserFile == "" {
		cfg.UserFile = defaultUserFile
	}

	if cfg.UserFile != "" {
		if fileExists(cfg.UserFile) {
			userMap, err := readYamlFile(cfg.UserFile)
			if err != nil {
				return err
			}
			user, err := UserConfigFromMap(userMap)
			if err != nil {
				return err
			}
			cfg.User = user
		}
	}

	if cfg.User != nil {
		// Transform deprecated user fields.
		if cfg.User.DeprecatedServices != nil {
			cfg.Warn("User configuration 'services' is deprecated in favor of 'modules'.")
			if cfg.User.Modules == nil {
				cfg.User.Modules = make(map[string]UserModuleConfig)
			}
			for k, v := range cfg.User.DeprecatedServices {
				if _, ok := cfg.User.Modules[k]; ok {
					cfg.Warn(fmt.Sprintf("User configuration 'services.%s' ignored since 'modules.%s' is present.", k, k))
				} else {
					cfg.User.Modules[k] = v
				}
			}
			cfg.User.DeprecatedServices = nil
		}
		if len(cfg.User.DeprecatedServicePreference) > 0 {
			cfg.Warn("User configuration 'service_preference' is deprecated in favor of 'module_order'.")
			cfg.User.ModuleOrder = append(cfg.User.DeprecatedServicePreference, cfg.User.ModuleOrder...)
			cfg.User.DeprecatedServicePreference = nil
		}

	}

	return nil
}

func loadModuleDefs(files []string) ([]*ModuleDef, error) {
	defs := make([]*ModuleDef, len(files))
	for i, file := range files {
		module := newModuleDef(file)
		msi, err := readYamlFile(file)
		if err != nil {
			return nil, err
		}
		err = mapToStruct(msi, module)
		if err != nil {
			return nil, err
		}
		// TODO: validate
		defs[i] = module
	}
	return defs, nil
}

func fileExists(file string) bool {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return false
	}
	return true
}

func mapToStruct(input interface{}, pointer interface{}) error {
	config := &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.StringToTimeDurationHookFunc(),
		// Make it clear to the user if something in the config
		// will have no effect.
		ErrorUnused: true,
		Result:      pointer,
		TagName:     "yaml",
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}
	return decoder.Decode(input)
}

func structToMap(input interface{}) (result map[string]interface{}, err error) {
	// The reflection in the yaml.Marshal call may panic.
	defer func() {
		if r := recover(); r != nil {
			result = nil
			err = fmt.Errorf("%s", r)
		}
	}()

	// We only do struct -> map for "config show" and testing, so keep it simple.
	var bs []byte
	bs, err = yaml.Marshal(input)
	if err != nil {
		return nil, err
	}

	var mii map[interface{}]interface{}
	err = yaml.Unmarshal(bs, &mii)
	if err != nil {
		return nil, err
	}

	return stringifyKeys(mii), nil
}

func readYamlFile(file string) (map[string]interface{}, error) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return parseYaml(content)
}

var yamlFileCache = make(map[string][]byte)

func readCachedYamlFile(file string) (map[string]interface{}, error) {
	if content, ok := yamlFileCache[file]; ok {
		return parseYaml(content)
	}
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	yamlFileCache[file] = content
	return parseYaml(content)
}

// parseYaml from `[]byte` and return a `map[string]interface{}`.
func parseYaml(content []byte) (map[string]interface{}, error) {
	var obj map[interface{}]interface{}
	err := yaml.Unmarshal(content, &obj)
	if err != nil {
		return nil, err
	}
	return stringifyKeys(obj), nil
}

func stringifyKeys(m map[interface{}]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(m))
	for k, v := range m {
		key := k.(string)

		if vm, ok := v.(map[interface{}]interface{}); ok {
			result[key] = stringifyKeys(vm)

		} else if vs, ok := v.([]interface{}); ok {
			slice := make([]interface{}, len(vs))
			for i, item := range vs {
				if m, ok := item.(map[interface{}]interface{}); ok {
					slice[i] = stringifyKeys(m)
				} else {
					slice[i] = item
				}
			}
			result[key] = slice

		} else {
			result[key] = v
		}
	}
	return result
}

func stringSlice(obj interface{}) ([]string, bool) {
	if slice, ok := obj.([]string); ok {
		return slice, true
	} else if slice, ok := obj.([]interface{}); ok {
		strings := make([]string, len(slice))
		for i, item := range slice {
			strings[i] = item.(string)
		}
		return strings, true
	}
	return nil, false
}
