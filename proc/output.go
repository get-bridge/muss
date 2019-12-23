package proc

import (
	"bytes"
	"os/exec"
)

// CmdOutput runs the specified command and returns it's stdout, stderr and any
// error returned from Run().
func CmdOutput(argv ...string) (string, string, error) {
	var stdout, stderr bytes.Buffer

	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	return string(bytes.TrimRight(stdout.Bytes(), "\n")),
		string(bytes.TrimRight(stderr.Bytes(), "\n")),
		err
}
