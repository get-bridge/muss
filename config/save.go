package config

import (
	"fmt"
	"log"
	"os"
	"path"
	"sync"

	yaml "gopkg.in/yaml.v2"
)

// Save writes out configuration files to disk.
func Save() {
	generateFiles(All())
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

func generateFiles(cfg *ProjectConfig) {
	// If there is not project config file, there is nothing to do.
	if cfg == nil {
		return
	}

	dc, err := cfg.ComposeConfig()
	if err != nil {
		log.Fatalln("Error creating docker-compose config:\n", err)
	}

	// If there's no compose config don't write the file.
	// If there is an existing docker-compose.yml it will get used
	// when muss delegates to docker-compose.
	if dc != nil {
		composeBytes := yamlDump(dc)

		if dc, err := os.Create(DockerComposeFile); err == nil {
			content := []byte(`#
# THIS FILE IS GENERATED!
#
# To add new service definition files edit ` + ProjectFile + `.
#
`)

			if cfg.UserFile != "" {
				content = append(content,
					[]byte(fmt.Sprintf("# To configure the services you want to use edit %v.\n#\n", cfg.UserFile))...)
			}

			content = append(content, []byte("\n---\n")...)
			content = append(content, composeBytes...)

			if _, err := dc.Write(content); err != nil {
				log.Fatalln("Error writing to file:", err)
			}
		} else {
			log.Fatalln("Failed to open file for writing:", err)
		}
	}

	var wg sync.WaitGroup
	files, _ := cfg.FilesToGenerate()
	for path, fn := range files {
		wg.Add(1)
		go func(path string, fn FileGenFunc) {
			defer wg.Done()
			if err := fn(path); err != nil {
				log.Fatalln("Failed to save file:", err)
			}
		}(path, fn)
	}
	wg.Wait()

	if err := loadEnvFromCmds(cfg.Secrets...); err != nil {
		log.Fatalln("Failed to load secrets:", err)
	}
}

func yamlDump(object map[string]interface{}) []byte {
	d, err := yaml.Marshal(object)
	if err != nil {
		log.Fatalln("yaml dump error:", err)
	}
	return d
}
