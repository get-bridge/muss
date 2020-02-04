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

// Save writes out configuration files to disk and returns any errors.
func Save() error {
	cfg, err := All()
	if err != nil {
		return err
	}
	return generateFiles(cfg)
}

// Assume that if a volume is pointing to an existing file they probably meant it.
func ensureExistsOrDir(file string) error {
	if _, err := os.Stat(file); err != nil {

		// If there was an error other not non-existence, return it.
		if !os.IsNotExist(err) {
			return err
		}

		if err := ensureDir(file); err != nil {
			// If we failed to make the directory because of a permission error
			// we do nothing and let docker deal with it.
			// This allows, e.g., "/dev/shm:/dev/shm" to work in Docker For Mac.
			if !os.IsPermission(err) {
				return err
			}
		}
	}

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

	fmt.Fprintf(stderr, "COMPOSE_FILE is set but does not contain muss target '%s'.\n", cfg.ComposeFilePath())
}

func yamlDump(object map[string]interface{}) ([]byte, error) {
	d, err := yaml.Marshal(object)
	if err != nil {
		return nil, err
	}
	return d, nil
}
