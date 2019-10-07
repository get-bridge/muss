package config

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func assertNotExist(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Lstat(path); err == nil {
		t.Fatalf("expected '%s' not to exist but it does", path)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat error: %s", err)
	}
}

func captureStderr(t *testing.T, f func()) string {
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

func readTestFile(t *testing.T, path string) string {
	t.Helper()
	content, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("error reading '%s': %s", path, err)
	}
	return string(content)
}

func tempdir(t *testing.T) string {
	t.Helper()
	tmp, err := ioutil.TempDir("", "muss-test")
	if err != nil {
		t.Fatalf("error creating temp dir: %s\n", err)
	}
	return tmp
}

func withTempDir(t *testing.T, f func(string)) {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current dir: %s\n", err)
	}

	xdgcache := os.Getenv("XDG_CACHE_HOME")
	home := os.Getenv("HOME")
	dir := tempdir(t)

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
