package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"

	"github.com/mitchellh/mapstructure"
	yaml "gopkg.in/yaml.v2"
)

// SetConfig sets the global config to the provided arg
// (used internally for testing).
func SetConfig(cfg map[string]interface{}) error {
	if cfg == nil {
		project = nil
		return nil
	}

	var err error
	project, err = prepare(cfg)
	if err != nil {
		return fmt.Errorf("Failed to process config: %w", err)
	}

	return nil
}

func load() error {
	if cfgFile := os.Getenv("MUSS_FILE"); cfgFile != "" {
		ProjectFile = cfgFile
	}
	UserFile = os.Getenv("MUSS_USER_FILE")

	// If there is no config file do the best you can
	// (allow muss to wrap docker-compose without a config file).
	if _, err := os.Stat(ProjectFile); err != nil && os.IsNotExist(err) {
		return nil
	}

	object, err := readYamlFile(ProjectFile)
	if err != nil {
		return fmt.Errorf("Failed to read config file '%s': %w", ProjectFile, err)
	}

	return SetConfig(object)
}

func prepare(object map[string]interface{}) (*ProjectConfig, error) {
	prepared, err := NewProjectConfigFromMap(object)
	if err != nil {
		return nil, err
	}

	loaded, err := loadServiceDefs(prepared.ServiceFiles)
	if err != nil {
		return nil, err
	}
	prepared.ServiceDefinitions = append(prepared.ServiceDefinitions, loaded...)

	if UserFile != "" {
		prepared.UserFile = UserFile
	}

	if prepared.UserFile != "" {
		if fileExists(prepared.UserFile) {
			userMap, err := readYamlFile(prepared.UserFile)
			if err != nil {
				return nil, err
			}
			user, err := UserConfigFromMap(userMap)
			if err != nil {
				return nil, err
			}
			prepared.User = user
		}
	}

	return prepared, nil
}

func loadServiceDefs(files []string) ([]ServiceDef, error) {
	defs := make([]ServiceDef, len(files))
	for i, file := range files {
		service, err := readYamlFile(file)
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
	// TODO hack around mapstructure saving structs
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
