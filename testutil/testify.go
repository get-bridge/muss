package testutil

import (
	"fmt"
	"os"
	"testing"
)

// functions modified (same interface) from an unreleased version of testify

// NoDirExists checks whether a directory does not exist in the given path.
// It fails if the path points to an existing _directory_ only.
func NoDirExists(t *testing.T, path string, msgAndArgs ...interface{}) bool {
	t.Helper()
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return true
		}
		return true
	}
	if !info.IsDir() {
		return true
	}
	t.Fatal(fmt.Sprintf("directory %q exists", path))
	return false
}

// NoFileExists checks whether a file does not exist in a given path. It fails
// if the path points to an existing _file_ only.
func NoFileExists(t *testing.T, path string, msgAndArgs ...interface{}) bool {
	t.Helper()
	info, err := os.Lstat(path)
	if err != nil {
		return true
	}
	if info.IsDir() {
		return true
	}
	t.Fatal(fmt.Sprintf("file %q exists", path))
	return false
}
