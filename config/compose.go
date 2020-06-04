package config

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
)

// FileGenFunc is a function that takes a string (file path)
// and generates the file at that path (returning an error).
type FileGenFunc func(string) error

// FileGenMap is just a map of the file path to its FileGenFunc.
type FileGenMap map[string]FileGenFunc

func (cfg *ProjectConfig) loadComposeConfig() error {
	if cfg.composeConfig == nil {
		if len(cfg.ModuleDefinitions) == 0 {
			dcc, err := cfg.loadStaticComposeConfig()
			if err != nil {
				return err
			}
			cfg.composeConfig = dcc
		} else {
			err := cfg.parseModuleDefinitions()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// ComposeConfig returns a map[string]interface ready to be yaml dumped as a
// docker-compose.yml file.
func (cfg *ProjectConfig) ComposeConfig() (map[string]interface{}, error) {
	if err := cfg.loadComposeConfig(); err != nil {
		return nil, err
	}
	return cfg.composeConfig, nil
}

// FilesToGenerate returns a FileGenMap of files to write.
func (cfg *ProjectConfig) FilesToGenerate() (FileGenMap, error) {
	if err := cfg.loadComposeConfig(); err != nil {
		return nil, err
	}
	return cfg.filesToGenerate, nil
}

// ComposeFilePath returns the path of the target compose file.
func (cfg *ProjectConfig) ComposeFilePath() string {
	if cfg != nil && cfg.ComposeFile != "" {
		return cfg.ComposeFile
	}
	return "docker-compose.yml"
}

// parseModuleDefinitions iterates the ProjectConfig.ModuleDefinitions
// to build up and store the docker compose map, file gen map, and secrets
// on the ProjectConfig value.
// If an error is returned no changes will be made to the value.
func (cfg *ProjectConfig) parseModuleDefinitions() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	// Setup a base to merge things onto.
	dcc := map[string]interface{}{
		"version": "3.7", // latest
	}
	files := make(FileGenMap)
	secrets := make([]envLoader, 0)

	for _, module := range cfg.ModuleDefinitions {
		servconf, err := module.chooseConfig(cfg)
		if err != nil {
			return err
		}

		secretsToParse := make([]map[string]interface{}, 0)
		if s, ok := servconf["secrets"]; ok {

			if mapsi, ok := s.(map[string]interface{}); ok {
				for varname, spec := range mapsi {
					if val, ok := spec.(map[string]interface{}); ok {
						secretsToParse = append(secretsToParse, mapMerge(map[string]interface{}{"varname": varname}, val))
					} else {
						return errors.New("secret spec must be a map")
					}
				}
			} else if slice, ok := s.([]interface{}); ok {
				for _, spec := range slice {
					if val, ok := spec.(map[string]interface{}); ok {
						secretsToParse = append(secretsToParse, val)
					} else {
						return errors.New("secret spec must be a map")
					}
				}
			}

			for _, spec := range secretsToParse {
				parsed, err := parseSecret(cfg, spec)
				if err != nil {
					return err
				}
				secrets = append(secrets, parsed)
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
					return err
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

	if yaml, err := cfg.composeFileBytes(dcc); err == nil {
		files[cfg.ComposeFilePath()] = fileGeneratorWithContent(yaml)
	} else {
		return err
	}

	// If we haven't returned any errors it's safe to update the value.

	cfg.composeConfig = dcc
	cfg.filesToGenerate = files
	cfg.Secrets = append(cfg.Secrets, secrets...)

	return nil
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
				prepare[path.Clean(source)] = attemptEnsureMountPointExists
			}
			// If this is a volume that will be mounted beneath the current
			// dir ensure the child dir exists.
			if len(wdTargets) > 0 {
				for wdTarget, wdSource := range wdTargets {
					if strings.HasPrefix(target, wdTarget) {
						subdir := path.Clean(strings.Replace(target, wdTarget, wdSource, 1))
						// Assume current dir is already a dir.
						if subdir != "." {
							prepare[subdir] = attemptEnsureMountPointExists
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

		// Break the reference for any maps that we copy over.
		if vmap, ok := v.(map[string]interface{}); ok {
			result[k] = mapMerge(map[string]interface{}{}, vmap)
		} else {
			result[k] = v
		}
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

func (cfg *ProjectConfig) composeFileBytes(dcc map[string]interface{}) ([]byte, error) {
	yamlBytes, err := yamlDump(dcc)
	if err != nil {
		return nil, err
	}

	content := []byte(`#
# THIS FILE IS GENERATED!
#
# To add new module definition files edit ` + cfg.ProjectFile + `.
#
`)

	if cfg.UserFile != "" {
		content = append(content,
			[]byte(fmt.Sprintf("# To configure the modules you want to use edit %v.\n#\n", cfg.UserFile))...)
	}

	content = append(content, []byte("\n---\n")...)
	content = append(content, yamlBytes...)

	return content, nil
}

func (cfg *ProjectConfig) loadStaticComposeConfig() (map[string]interface{}, error) {
	m, err := readYamlFile(cfg.ComposeFilePath())
	if err != nil {
		// Don't abort config loading if this doesn't exist.
		if !os.IsNotExist(err) {
			return nil, err
		}
	}
	return m, nil
}
