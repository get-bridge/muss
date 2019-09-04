package config

// DockerComposeFile is the path to the docker-compose file that will be
// generated.
var DockerComposeFile = "docker-compose.yml"

// DockerComposeConfig returns an object ready to be yaml-dumped as a
// docker-compose file.
func DockerComposeConfig(config ProjectConfig) map[string]interface{} {
	// Setup a base to merge things onto.
	dcc := map[string]interface{}{
		"version":  "3.7", // latest
		"volumes":  map[string]interface{}{},
		"services": map[string]interface{}{},
	}

	var servdefs []ServiceDef
	if val, ok := config["service_definitions"].([]ServiceDef); ok {
		servdefs = val
	} else if val, ok := config["service_definitions"].([]interface{}); ok {
		servdefs = make([]ServiceDef, len(val))
		for i, s := range val {
			servdefs[i] = s.(map[string]interface{})
		}
	}

	for _, service := range servdefs {
		servconf := serviceConfig(config, service)

		dcc = mapMerge(dcc, servconf)
	}
	return dcc
}

func serviceConfig(config map[string]interface{}, service ServiceDef) map[string]interface{} {
	serviceName := service["name"].(string)
	serviceConfigs := service["configs"].(map[string]interface{})
	options := mapKeys(serviceConfigs)

	var result map[string]interface{}
	userChoice := ""
	// TODO: check user.services.config for a name
	if userChoice != "" {
		result = serviceConfigs[serviceName].(map[string]interface{})
	} else if len(options) == 1 {
		result = serviceConfigs[options[0]].(map[string]interface{})
	} else {

		// TODO: merge user config
		order := config["default_service_preference"].([]interface{})
		for _, o := range order {
			if found, ok := serviceConfigs[o.(string)]; ok {
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
	return result
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
