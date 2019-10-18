package config

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/mitchellh/mapstructure"
	yaml "gopkg.in/yaml.v2"
)

// SetConfig sets the global config to the provided arg
// (used internally for testing).
func SetConfig(cfg ProjectConfig) {
	if cfg == nil {
		project = nil
		return
	}

	prepared, err := prepare(cfg)
	if err != nil {
		log.Fatalln("Failed to process config:", err)
	}
	project = prepared
}

func load() {
	// If there is no config file do the best you can
	// (allow muss to wrap docker-compose without a config file).
	if _, err := os.Stat(ProjectFile); err != nil && os.IsNotExist(err) {
		return
	}

	object, err := readYamlFile(ProjectFile)
	if err != nil {
		log.Fatalf("Failed to read config file '%s': %s\n", ProjectFile, err)
	}
	SetConfig(object)
}

func prepare(object map[string]interface{}) (ProjectConfig, error) {
	prepared := object

	if files, ok := stringSlice(object["service_files"]); ok {
		loaded, err := loadServiceDefs(files)
		if err != nil {
			return nil, err
		}
		prepared = mapMerge(prepared, map[string]interface{}{"service_definitions": loaded})

		// TODO: else: mention that they might want service_files?
	}

	if UserFile != "" {
		prepared["user_file"] = UserFile
	}

	if file, ok := prepared["user_file"].(string); ok {
		if fileExists(file) {
			user, err := readYamlFile(file)
			if err != nil {
				return nil, err
			}
			prepared["user"] = user
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

func subMap(cfg ProjectConfig, key string) map[string]interface{} {
	if val, ok := cfg[key].(map[string]interface{}); ok {
		return val
	}
	return map[string]interface{}{}
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
