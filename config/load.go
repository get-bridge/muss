package config

import (
	"io/ioutil"
	"log"

	yaml "gopkg.in/yaml.v2"
)

// SetConfig sets the global config to the provided arg
// (used internally for testing).
func SetConfig(cfg ProjectConfig) {
	prepared, err := prepare(cfg)
	if err != nil {
		log.Fatalln("Failed to process config:", err)
	}
	project = prepared
}

func load() {
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
	if slice, ok := obj.([]interface{}); ok {
		strings := make([]string, len(slice))
		for i, item := range slice {
			strings[i] = item.(string)
		}
		return strings, true
	}
	return nil, false
}
