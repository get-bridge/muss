package config

import (
	"fmt"
	"os"
	"regexp"
)

func expand(s string) string {
	return os.Expand(s, shellVarExpand)
}

func expandWarnOnEmpty(s string) string {
	return os.Expand(s, shellVarExpandWarnOnEmpty)
}

func shellVarExpand(spec string) string {
	re := regexp.MustCompile(`^([_a-zA-Z][_a-zA-Z0-9]*)(?:(:?[-?])(.*))?$`)
	match := re.FindStringSubmatch(spec)
	requiredErrorSpec := "Variable '%s' is required: %s"

	if len(match) > 0 {
		name, op, value := match[1], match[2], match[3]
		switch op {
		case "": // no operator
			return os.Getenv(name)
		case ":-": // null or unset
			env := os.Getenv(name)
			if env == "" {
				return value
			}
			return env
		case "-": // unset
			env, ok := os.LookupEnv(name)
			if !ok {
				return value
			}
			return env
		case ":?": // error if null or unset
			env := os.Getenv(name)
			if env == "" {
				panic(fmt.Sprintf(requiredErrorSpec, name, value))
			}
			return env
		case "?": // error if null
			env, ok := os.LookupEnv(name)
			if !ok {
				panic(fmt.Sprintf(requiredErrorSpec, name, value))
			}
			return env
		}
	}

	panic(fmt.Sprintf("Invalid interpolation format: '${%s}'", spec))
}

func shellVarExpandWarnOnEmpty(spec string) string {
	val := shellVarExpand(spec)
	if val == "" {
		fmt.Fprintf(os.Stderr, "${%s} is blank\n", spec)
	}
	return val
}
