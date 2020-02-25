package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"

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

	loaded, err := loadServiceDefs(cfg.ServiceFiles)
	if err != nil {
		return err
	}
	cfg.ServiceDefinitions = append(cfg.ServiceDefinitions, loaded...)

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

	return nil
}

func loadServiceDefs(files []string) ([]*ServiceDef, error) {
	defs := make([]*ServiceDef, len(files))
	for i, file := range files {
		service := newServiceDef()
		msi, err := readYamlFile(file)
		if err != nil {
			return nil, err
		}
		err = mapToStruct(msi, service)
		if err != nil {
			return nil, err
		}
		// TODO: validate
		defs[i] = service
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

func structToMap(input interface{}) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := mapToStruct(input, &result); err != nil {
		return nil, err
	}

	// Descend into map as mapstructure is only doing the top level.
	for key, value := range result {
		switch reflect.TypeOf(value).Kind() {
		case reflect.Ptr:
			fallthrough
		case reflect.Struct:
			fallthrough
		case reflect.Map:
			mapValue, err := structToMap(value)
			if err != nil {
				return nil, err
			}
			result[key] = mapValue
		case reflect.Slice:
			sliceVal := reflect.ValueOf(value)
			length := sliceVal.Len()
			// TODO: more than ptr
			if length > 0 && sliceVal.Index(0).Kind() == reflect.Ptr {
				if sliceVal.Index(0).CanInterface() {
					mapped := make([]interface{}, length)
					for i := 0; i < length; i++ {
						mapValue, err := structToMap(sliceVal.Index(i).Interface())
						if err != nil {
							return nil, err
						}
						mapped[i] = mapValue
					}
					result[key] = mapped
				}
			}
		}
	}
	return result, nil
}

func readYamlFile(file string) (map[string]interface{}, error) {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
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
