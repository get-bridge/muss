package config

import (
	"log"
	"os"
	"path"

	yaml "gopkg.in/yaml.v2"
)

// Save writes out configuration files to disk.
func Save() {
	generateFiles(All())
}

// Assume that if a volume is pointing to an existing file they probably meant it.
func ensureExistsOrDir(file string) error {
	if _, err := os.Stat(file); err != nil {
		if os.IsNotExist(err) {
			return ensureDir(file)
		}
		return err
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

func generateFiles(cfg ProjectConfig) {
	// If there is not project config file, there is nothing to do.
	if cfg == nil {
		return
	}

	dc, files, err := DockerComposeFiles(cfg)
	if err != nil {
		log.Fatalln("Error creating docker-compose config:\n", err)
	}
	composeBytes := yamlDump(dc)

	if dc, err := os.Create(DockerComposeFile); err == nil {
		content := []byte(`#
# THIS FILE IS GENERATED!
#
# To add new service definition files edit ` + ProjectFile + `.
#
`)

		if file, ok := cfg["user_file"]; ok {
			content = append(content,
				[]byte("# To configure the services you want to use edit "+file.(string)+".\n#\n")...)
		}

		content = append(content, []byte("\n---\n")...)
		content = append(content, composeBytes...)

		if _, err := dc.Write(content); err != nil {
			log.Fatalln("Error writing to file:", err)
		}
	} else {
		log.Fatalln("Failed to open file for writing:", err)
	}

	// TODO: run these concurrently and bubble errors.
	for path, fn := range files {
		if err := fn(path); err != nil {
			log.Fatalln("Failed to save file:", err)
		}
	}

	// TODO: run these concurrently and bubble errors.
	for _, secret := range projectSecrets {
		if err := secret.load(); err != nil {
			log.Fatalln("Failed to get secret:", err)
		}
	}
}

func yamlDump(object map[string]interface{}) []byte {
	d, err := yaml.Marshal(object)
	if err != nil {
		log.Fatalln("yaml dump error:", err)
	}
	return d
}
