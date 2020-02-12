package testutil

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

// Tempdir returns a tempdir or calls t.Fatal.
func Tempdir(t *testing.T) string {
	t.Helper()
	tmp, err := ioutil.TempDir("", "muss-test")
	if err != nil {
		t.Fatalf("error creating temp dir: %s\n", err)
	}
	return tmp
}

// WithTempDir creates a tempdir, passes it to its function argument,
// then cleans up the tempdir when done.
func WithTempDir(t *testing.T, f func(string)) {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current dir: %s\n", err)
	}

	xdgcache := os.Getenv("XDG_CACHE_HOME")
	home := os.Getenv("HOME")
	dir := Tempdir(t)

	os.Setenv("HOME", path.Join(dir, "test-home"))
	os.Setenv("XDG_CACHE_HOME", path.Join(dir, "test-cache"))
	os.Chdir(dir)

	defer func() {
		os.Setenv("HOME", home)
		os.Setenv("XDG_CACHE_HOME", xdgcache)
		os.Chdir(cwd)
		os.RemoveAll(dir)
	}()

	f(dir)
}
