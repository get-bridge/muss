package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	multierror "github.com/hashicorp/go-multierror"
	yaml "gopkg.in/yaml.v2"
)

// Save writes out the generated config files to disk.
func (cfg *ProjectConfig) Save() error {
	return generateFiles(cfg)
}

// Assume that if a volume is pointing to an existing file they probably meant it.
func attemptEnsureMountPointExists(file string) error {
	// MkdirAll will only proceed if the path doesn't already exist (file or dir).
	ensureDir(file)
	// Ignore any errors (permission or special FS issues) and let
	// docker-compose / docker deal with anything that doesn't exist.
	return nil
}

func ensureDir(file string) error {
	return os.MkdirAll(file, 0777)
}

func ensureFile(file string) error {
	if err := ensureDir(path.Dir(file)); err != nil {
		return err
	}
	if stat, err := os.Stat(file); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else if stat.IsDir() {
		// if empty remove it, else complain
		if err := os.Remove(file); err != nil {
			return err
		}
	} else {
		// already a file.
		return nil
	}

	return touch(file)
}

func touch(file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	f.Close()
	return nil
}

func fileGeneratorWithContent(content []byte) FileGenFunc {
	return func(file string) error {
		return ioutil.WriteFile(file, content, 0666)
	}
}

func generateFiles(cfg *ProjectConfig) error {
	// If there is not project config file, there is nothing to do.
	if cfg == nil {
		return nil
	}

	files, err := cfg.FilesToGenerate()
	if err != nil {
		return fmt.Errorf("Error creating docker-compose config: %w", err)
	}

	cfg.checkComposeFileVar()

	var wg sync.WaitGroup
	var errs *multierror.Error
	var errMux sync.Mutex

	for path, fn := range files {
		wg.Add(1)
		go func(path string, fn FileGenFunc) {
			defer wg.Done()
			if err := fn(path); err != nil {
				errMux.Lock()
				errs = multierror.Append(errs, err)
				errMux.Unlock()
			}
		}(path, fn)
	}
	wg.Wait()

	if errs != nil && len(errs.Errors) > 0 {
		if len(errs.Errors) == 1 {
			return errs.Errors[0]
		}
		return errs.ErrorOrNil()
	}

	return cfg.LoadEnv()

}

func (cfg *ProjectConfig) checkComposeFileVar() {
	cfile, ok := os.LookupEnv("COMPOSE_FILE")
	if !ok {
		return
	}

	sep := string(os.PathListSeparator)
	if val, ok := os.LookupEnv("COMPOSE_PATH_SEPARATOR"); ok {
		sep = val
	}
	paths := strings.Split(cfile, sep)
	for _, path := range paths {
		if filepath.Base(path) == cfg.ComposeFilePath() {
			return
		}
	}

	cfg.Warn(fmt.Sprintf("COMPOSE_FILE is set but does not contain muss target '%s'.", cfg.ComposeFilePath()))
}

func yamlDump(object map[string]interface{}) ([]byte, error) {
	d, err := yaml.Marshal(object)
	if err != nil {
		return nil, err
	}
	return d, nil
}
