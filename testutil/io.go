package testutil

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// CaptureStderr sets os.Stderr to a tempfile, runs the provided function,
// and returns a string of what was written to Stderr during that function.
func CaptureStderr(t *testing.T, f func()) string {
	t.Helper()

	// os.Stderr must be a *os.File
	fake, err := ioutil.TempFile("", "muss-stderr")
	if err != nil {
		t.Fatalf("failed to create temp file for stderr: %s", err)
	}

	orig := os.Stderr
	os.Stderr = fake

	defer func() {
		// reset this even if f() panics
		os.Stderr = orig

		os.Remove(fake.Name())
	}()

	f()

	// reset this as early as possible
	os.Stderr = orig

	fake.Close()
	content, err := ioutil.ReadFile(fake.Name())
	if err != nil {
		t.Fatalf("failed to read temp stderr: %s", err)
	}
	return string(content)
}

// ReadFile returns the contents of the file as a string
// and calls t.Fatal if there is an error.
func ReadFile(t *testing.T, path string) string {
	t.Helper()
	content, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("error reading '%s': %s", path, err)
	}
	return string(content)
}

// TempFile returns a TempFile via ioutil and calls t.Fatal on error.
func TempFile(t *testing.T, dir, pattern string) *os.File {
	tmpfile, err := ioutil.TempFile(dir, pattern)
	if err != nil {
		t.Fatalf("error making tempfile '%s/%s': %s", dir, pattern, err)
	}
	return tmpfile
}

// WriteFile writes the file with the provided string
// and calls t.Fatal if there is an error.
func WriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatalf("error making parent for %s: %s", path, err)
	}
	if err := ioutil.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("error reading '%s': %s", path, err)
	}
}
