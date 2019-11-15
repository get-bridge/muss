package config

import (
	"fmt"
	"path"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
)

// FileGenFunc is a function that takes a string (file path)
// and generates the file at that path (returning an error).
type FileGenFunc func(string) error

// FileGenMap is just a map of the file path to its FileGenFunc.
type FileGenMap map[string]FileGenFunc

// DockerComposeFile is the path to the docker-compose file that will be
// generated.
var DockerComposeFile = "docker-compose.yml"

// GenerateDockerComposeConfig returns an object ready to be yaml-dumped as a
// docker-compose file (or an error).
func GenerateDockerComposeConfig(config *ProjectConfig) (DockerComposeConfig, error) {
	dcc, _, err := GenerateDockerComposeFiles(config)
	return dcc, err
}

// GenerateDockerComposeFiles returns a map of the docker-compose config,
// a map representing supplementary files, and an error.
// The files value is a `map[string]func(string) error` where the key is
// the file path and the value is a function that takes the path argument
// and writes the file or errors.
func GenerateDockerComposeFiles(cfg *ProjectConfig) (dcc DockerComposeConfig, files FileGenMap, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	files = make(FileGenMap)

	// If there are no service defs there's nothing for us to write.
	if len(cfg.ServiceDefinitions) == 0 {
		return nil, files, nil
	}

	// Setup a base to merge things onto.
	dcc = NewDockerComposeConfig()

	for _, service := range cfg.ServiceDefinitions {
		servconf, err := serviceConfig(cfg, service)
		if err != nil {
			return nil, nil, err
		}

		secretsToParse := make([]map[string]interface{}, 0)
		if s, ok := servconf["secrets"]; ok {

			if mapsi, ok := s.(map[string]interface{}); ok {
				for varname, spec := range mapsi {
					if val, ok := spec.(map[string]interface{}); ok {
						secretsToParse = append(secretsToParse, mapMerge(map[string]interface{}{"varname": varname}, val))
					} else {
						return nil, nil, fmt.Errorf("secret spec must be a map")
					}
				}
			} else if slice, ok := s.([]interface{}); ok {
				for _, spec := range slice {
					if val, ok := spec.(map[string]interface{}); ok {
						secretsToParse = append(secretsToParse, val)
					} else {
						return nil, nil, fmt.Errorf("secret spec must be a map")
					}
				}
			}

			for _, spec := range secretsToParse {
				parsed, err := parseSecret(cfg, spec)
				if err != nil {
					return nil, nil, err
				}
				// TODO make it obvious we are modifying func arg
				cfg.Secrets = append(cfg.Secrets, parsed)
			}

			delete(servconf, "secrets")
		}

		dcc = mapMerge(dcc, servconf)
	}

	if cfg.User != nil && cfg.User.Override != nil {
		dcc = mapMerge(dcc, cfg.User.Override)
	}

	// Iterate over each service to remove any muss extensions
	// and do any necessary preparations.
	if services, ok := (dcc["services"]).(map[string]interface{}); ok {
		for name, si := range services {
			if service, ok := si.(map[string]interface{}); ok {

				bindvols, err := prepareVolumes(service)
				if err != nil {
					return nil, nil, err
				}
				for path, fn := range bindvols {
					files[path] = fn
				}

				if !isValidService(service) {
					delete(services, name)
				}

			}
		}
	}

	return dcc, files, nil
}

func serviceConfig(config *ProjectConfig, service ServiceDef) (ServiceConfig, error) {
	serviceName := service["name"].(string)
	serviceConfigs := service["configs"].(map[string]interface{})
	options := mapKeys(serviceConfigs)
	result := NewServiceConfig()

	// Check if user configured this service specifically.
	userChoice := ""
	if config.User != nil {
		if userserv, ok := config.User.Services[serviceName]; ok {
			if userserv.Disabled {
				return result, nil
			}

			userChoice = userserv.Config
			if userChoice != "" {
				if _, ok := serviceConfigs[userChoice]; !ok {
					return nil, fmt.Errorf("Config '%s' for service '%s' does not exist", userChoice, serviceName)
				}
			}
		}
	}

	if userChoice != "" {
		// If user chose specifically, use it.
		result = serviceConfigs[userChoice].(map[string]interface{})
	} else if len(options) == 1 {
		// If there is only one option, use it.
		result = serviceConfigs[options[0]].(map[string]interface{})
	} else {
		// To determine which config option to use we can build a list...
		// starting with any user configured preference...
		var order []string
		if config.User != nil {
			order = config.User.ServicePreference
		} else {
			order = []string{}
		}
		// followed by any project defaults...
		order = append(order, config.DefaultServicePreference...)

		// then iterate and use the first preference that this service defines.
		for _, o := range order {
			if found, ok := serviceConfigs[o]; ok {
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
			base = mapMerge(base, serviceConfigs[i.(string)].(map[string]interface{}))
		}
		result = mapMerge(base, result)
	}
	return result, nil
}

func isValidService(service map[string]interface{}) bool {
	if _, ok := service["build"]; ok {
		return true
	}
	if _, ok := service["image"]; ok {
		return true
	}
	return false
}

// Iterate over the volumes for this service looking for files (muss extension)
// or directories and pre-create them when we can to avoid the issues of
// docker creating a directory where we wanted a file
// or creating a directory now owned by root when the user should own it.
// NOTE: This currently assumes unix-like (forward slash) paths.
func prepareVolumes(service map[string]interface{}) (FileGenMap, error) {
	prepare := make(FileGenMap)
	wdTargets := make(map[string]string)

	// First try to find if we are bind-mounting the current dir (or child dirs).
	if err := iterateBindMounts(service, func(source, target string, volume map[string]interface{}) {
		// Don't clean the source (yet) because it's the leading "./" that
		// determines if this is a bind mount.
		if source == "." || strings.HasPrefix(source, "./") {
			wdTargets[endWithPathSep(target)] = endWithPathSep(source)
		}
	}); err != nil {
		return nil, err
	}

	if err := iterateBindMounts(service, func(source, target string, volume map[string]interface{}) {
		// > NOTE: File must exist, else "It is always created as a directory".
		// > https://docs.docker.com/storage/bind-mounts/#differences-between--v-and---mount-behavior
		if file, ok := volume["file"].(bool); ok && file {
			prepare[path.Clean(source)] = ensureFile
			// docker-compose will abort if it gets a key it doesn't recognize.
			delete(volume, "file")
		} else {
			// If we are bind mounting a dir ensure it exists
			// else docker will create it and it will be owned by root.
			// Test for "/" or "./something" (ignore "." and "./")
			if strings.HasPrefix(source, "/") || (len(source) > 2 && source[0:2] == "./") {
				prepare[path.Clean(source)] = ensureExistsOrDir
			}
			// If this is a volume that will be mounted beneath the current
			// dir ensure the child dir exists.
			if len(wdTargets) > 0 {
				for wdTarget, wdSource := range wdTargets {
					if strings.HasPrefix(target, wdTarget) {
						subdir := path.Clean(strings.Replace(target, wdTarget, wdSource, 1))
						// Assume current dir is already a dir.
						if subdir != "." {
							prepare[subdir] = ensureExistsOrDir
						}
					}
				}
			}
		}
	}); err != nil {
		return nil, err
	}

	return prepare, nil
}

func endWithPathSep(s string) string {
	return path.Clean(s) + "/"
}

func iterateBindMounts(service map[string]interface{}, f func(string, string, map[string]interface{})) error {
	if volumes, ok := service["volumes"].([]interface{}); ok {
		for _, volume := range volumes {
			if v, ok := volume.(map[string]interface{}); ok {
				if v["type"] == "bind" {
					source, err := homedir.Expand(expand(v["source"].(string)))
					if err != nil {
						return err
					}
					f(source, v["target"].(string), v)
				}
			} else if v, ok := volume.(string); ok {
				expanded, err := homedir.Expand(expand(v))
				if err != nil {
					return err
				}
				parts := strings.Split(expanded, ":")
				source, target := parts[0], parts[1]
				// We could fake the volume long-syntax map here but we don't currently need it.
				f(source, target, nil)
			}
		}
	}

	return nil
}

var keysToOverwrite = []string{"entrypoint", "command"}

func mapMerge(target map[string]interface{}, source map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(target)+len(source))
	for k, v := range target {
		result[k] = v
	}
	for k, v := range source {
		if !mapMergeOverwrites(k) {
			if current, ok := result[k]; ok {
				if m, ok := current.(map[string]interface{}); ok {
					result[k] = mapMerge(m, v.(map[string]interface{}))
					continue
				} else if s, ok := current.([]interface{}); ok {
					vs := v.([]interface{})
					tmp := make([]interface{}, 0, len(s)+len(vs))
					tmp = append(tmp, s...)
					tmp = append(tmp, vs...)
					result[k] = tmp
					continue
				}
			}
		}
		result[k] = v
	}
	return result
}

func mapMergeOverwrites(k string) bool {
	for _, o := range keysToOverwrite {
		if o == k {
			return true
		}
	}
	return false
}

func mapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		if k[0:1] != "_" {
			keys = append(keys, k)
		}
	}
	return keys
}
