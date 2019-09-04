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
	composeBytes := yamlDump(DockerComposeConfig(cfg))

	if dc, err := os.Create(DockerComposeFile); err == nil {
		dc.Write(append([]byte(`#
# THIS FILE IS GENERATED!
#
# To add new service definition files edit `+ProjectFile+`.
#
---
`), composeBytes...))
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
