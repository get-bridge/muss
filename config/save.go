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

func ensureFile(file string) error {
	if err := os.MkdirAll(path.Dir(file), 0777); err != nil {
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

	if file, err := os.Create(file); err == nil {
		file.Close()
	} else {
		return err
	}

	return nil
}

func generateFiles(cfg ProjectConfig) {
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

	for path, fn := range files {
		if err := fn(path); err != nil {
			log.Fatalln("Failed to save file:", err)
		}
	}

	// TODO: secrets files
}

func yamlDump(object map[string]interface{}) []byte {
	d, err := yaml.Marshal(object)
	if err != nil {
		log.Fatalln("yaml dump error:", err)
	}
	return d
}
