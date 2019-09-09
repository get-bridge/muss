package config

import (
	"log"
	"os"

	yaml "gopkg.in/yaml.v2"
)

// Save writes out configuration files to disk.
func Save() {
	generateFiles(All())
}

func generateFiles(cfg ProjectConfig) {
	dc, err := DockerComposeConfig(cfg)
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

	// TODO: secrets files
}

func yamlDump(object map[string]interface{}) []byte {
	d, err := yaml.Marshal(object)
	if err != nil {
		log.Fatalln("yaml dump error:", err)
	}
	return d
}
