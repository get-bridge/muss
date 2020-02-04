package config

import (
	"io"
	"os"
)

var stderr io.Writer

func init() {
	stderr = os.Stderr
}

// SetStderr allows tests to overwrite the destination for warnings.
func SetStderr(writer io.Writer) {
	stderr = writer
}
