package config

import (
	"fmt"
	"io"
	"text/template"

	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"

	rootcmd "gerrit.instructure.com/muss/cmd"
	"gerrit.instructure.com/muss/config"
)

var format = "{{ yaml . }}"

func newShowCommand(cfg *config.ProjectConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "show",
		Short: "Show muss config",
		Long: `
Useful for debugging and scripting.
The project configuration is the template context.

Additional functions available to template:
  compose: the docker compose config
  yaml: format arg as yaml

Template examples:

  # Show all the volumes in the current config:
  '{{ range compose.services }}{{ range .volumes }}{{ . }}{{ "\n" }}{{ end }}{{ end }}'

  # Show the names of each configured service:
  '{{ range $k, $v := compose.services }}{{ $k }}{{ "\n" }}{{ end }}'

  # Show all the options for service configs:
  '{{ range .service_definitions }}{{ range $k, $v := .configs }}{{ $k }}{{ "\n" }}{{ end }}{{end }}'
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := processTemplate(format, cfg, cmd.OutOrStdout())
			return rootcmd.QuietErrorOrNil(err)
		},
	}

	cmd.Flags().StringVar(&format, "format", format, "Format the output using the given Go template")

	return cmd
}

func processTemplate(format string, cfg *config.ProjectConfig, writer io.Writer) error {
	cfgMap, err := cfg.ToMap()
	if err != nil {
		return err
	}

	funcMap := template.FuncMap{
		"compose": func() map[string]interface{} {
			dc, err := cfg.ComposeConfig()
			if err != nil {
				panic(err)
			}
			return dc
		},
		"yaml": yamlToString,
		// for ease and consistency with compose...
		"project": func() map[string]interface{} {
			return cfgMap
		},
		"user": func() map[string]interface{} {
			if cfg.User != nil {
				cfgMap, err := cfg.User.ToMap()
				if err != nil {
					panic(err)
				}
				return cfgMap
			}
			return map[string]interface{}{}
		},
	}

	t, err := template.New("config").Funcs(funcMap).Parse(format)
	if err != nil {
		return err
	}
	return t.Execute(writer, cfgMap)
}

func yamlToString(object interface{}) string {
	bs, err := yaml.Marshal(object)
	if err != nil {
		// This error will be returned by t.Execute().
		panic(fmt.Errorf("unable to marshal object to YAML: %v", err))
	}
	return string(bs)
}

func init() {
	AddCommandBuilder(newShowCommand)
}
