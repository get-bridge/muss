package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeprecatedFieldsFromModuleToServiceRename(t *testing.T) {
	t.Run("old", func(t *testing.T) {

		cfg := newTestConfig(t, map[string]interface{}{
			"service_files":              []string{"../testdata/app.yml", "../testdata/store.yml"},
			"default_service_preference": []string{"repo", "registry"},
			"user": map[string]interface{}{
				"service_preference": []string{"registry", "repo"},
				"services": map[string]interface{}{
					"app": map[string]interface{}{
						"disabled": true,
					},
				},
			},
		})

		assert.Equal(t, []string{
			"Configuration 'service_files' is deprecated in favor of 'module_files'.",
			"Configuration 'default_service_preference' is deprecated in favor of 'default_module_order'.",
			"User configuration 'services' is deprecated in favor of 'modules'.",
			"User configuration 'service_preference' is deprecated in favor of 'module_order'.",
		}, cfg.Warnings, "deprecation warnings")

		moduleNames := func() []string {
			names := make([]string, len(cfg.ModuleDefinitions))
			for i, m := range cfg.ModuleDefinitions {
				names[i] = m.Name
			}
			return names
		}()

		assert.Equal(t, []string{"app", "store"}, moduleNames, "module defs initialized")

		assert.Equal(t, []string{
			"../testdata/app.yml",
			"../testdata/store.yml",
		}, cfg.ModuleFiles, "service_files -> module_files")

		assert.Equal(t, []string{"repo", "registry"}, cfg.DefaultModuleOrder, "default_service_preference -> default_module_order")

		assert.Equal(t, []string{"registry", "repo"}, cfg.User.ModuleOrder, "user service_preference -> module_order")
		assert.Equal(t, map[string]UserModuleConfig{
			"app": UserModuleConfig{Disabled: true},
		}, cfg.User.Modules, "user services -> modules")

		cfgMap, err := cfg.ToMap()
		assert.Nil(t, err)

		hasKey := func(m map[string]interface{}, k string) bool {
			_, ok := m[k]
			return ok
		}

		for _, key := range []string{"module_definitions", "module_files", "default_module_order"} {
			assert.True(t, hasKey(cfgMap, key), "map has "+key)
		}

		for _, key := range []string{"service_definitions", "service_files", "default_service_preference"} {
			assert.False(t, hasKey(cfgMap, key), "map does not have "+key)
		}

		userMap, ok := cfgMap["user"].(map[string]interface{})
		assert.True(t, ok)

		for _, key := range []string{"modules", "module_order"} {
			assert.True(t, hasKey(userMap, key), "user map has "+key)
		}

		for _, key := range []string{"services", "service_preference"} {
			assert.False(t, hasKey(userMap, key), "user map does not have "+key)
		}
	})

	t.Run("both", func(t *testing.T) {

		cfg := newTestConfig(t, map[string]interface{}{
			"service_definitions": []interface{}{
				map[string]interface{}{
					"name":    "sd",
					"configs": map[string]interface{}{},
				},
			},
			"module_definitions": []interface{}{
				map[string]interface{}{
					"name":    "md",
					"configs": map[string]interface{}{},
				},
			},
			"service_files":              []string{"../testdata/app.yml"},
			"module_files":               []string{"../testdata/store.yml"},
			"default_service_preference": []string{"repo"},
			"default_module_order":       []string{"registry"},
			"user": map[string]interface{}{
				"module_order":       []string{"c", "d"},
				"service_preference": []string{"a", "b"},
				"modules": map[string]interface{}{
					"app": map[string]interface{}{
						"config": "foo",
					},
				},
				"services": map[string]interface{}{
					"app": map[string]interface{}{
						"disabled": true,
						"config":   "bar",
					},
					"store": map[string]interface{}{
						"disabled": true,
					},
				},
			},
		})

		assert.Equal(t, []string{
			"Configuration 'service_definitions' is deprecated in favor of 'module_definitions'.",
			"Configuration 'service_files' is deprecated in favor of 'module_files'.",
			"Configuration 'default_service_preference' is deprecated in favor of 'default_module_order'.",
			"User configuration 'services' is deprecated in favor of 'modules'.",
			"User configuration 'services.app' ignored since 'modules.app' is present.",
			"User configuration 'service_preference' is deprecated in favor of 'module_order'.",
		}, cfg.Warnings, "deprecation warnings")

		moduleNames := func() []string {
			names := make([]string, len(cfg.ModuleDefinitions))
			for i, m := range cfg.ModuleDefinitions {
				names[i] = m.Name
			}
			return names
		}()

		assert.Equal(t, []string{"sd", "md", "app", "store"}, moduleNames, "service_definitions -> module_definitions")

		assert.Equal(t, []string{
			"../testdata/app.yml",
			"../testdata/store.yml",
		}, cfg.ModuleFiles, "service_files -> module_files")

		assert.Equal(t, []string{"repo", "registry"}, cfg.DefaultModuleOrder, "default_service_preference -> default_module_order")

		assert.Equal(t, []string{"a", "b", "c", "d"}, cfg.User.ModuleOrder, "user service_preference -> module_order")
		assert.Equal(t, map[string]UserModuleConfig{
			"app":   UserModuleConfig{Config: "foo"},
			"store": UserModuleConfig{Disabled: true},
		}, cfg.User.Modules, "user services -> modules")

	})
}
