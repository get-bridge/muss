package config

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecretCommands(t *testing.T) {
	var secretCmdPath string
	if dir, err := os.Getwd(); err != nil {
		t.Fatalf("failed to get working dir: %s", err)
	} else {
		secretCmdPath = path.Join(dir, "..", "testdata", "bin", "some-secret")
	}

	withTempDir(t, func(tmpdir string) {
		findCacheRoot()

		os.Setenv("MUSS_TEST_PASSPHRASE", "go test!")

		cfg := map[string]interface{}{
			"secrets": map[string]interface{}{
				"passphrase": "$MUSS_TEST_PASSPHRASE",
				"commands": map[string]interface{}{
					"some": map[string]interface{}{
						"exec": []string{secretCmdPath, "something"},
					},
				},
			},
		}

		varname := "MUSS_TEST_SECRET_VAR"
		os.Unsetenv(varname)

		secretSpec := map[string]interface{}{
			"some":    []string{"green"},
			"varname": varname,
		}

		logvarname := "MUSS_TEST_SECRET_LOG"
		os.Setenv(logvarname, "shhh")

		secret, err := parseSecret(cfg, secretSpec)
		if err != nil {
			t.Fatalf("error preparing secret env file: %s", err)
		}

		secretCacheFile := path.Join(secretDir, genFileName([]string{secretCmdPath, "something", "green"}))
		secretLog := "secret-log.txt"

		assertNotExist(t, secretCacheFile)

		os.Setenv(varname, "oops")
		testLoadSecret(t, secret)
		assert.Equal(t, os.Getenv(varname), "oops", "existing var not overwritten")
		assertNotExist(t, secretLog) // secret not called

		expSecret := "secret is [something green]"

		os.Unsetenv(varname)
		testLoadSecret(t, secret)
		assert.Equal(t, expSecret, os.Getenv(varname), "sets env var")
		assert.FileExists(t, secretCacheFile)
		assert.Equal(t, "shhh\n", readTestFile(t, secretLog), "secret called once")

		os.Setenv(logvarname, "again")

		os.Unsetenv(varname)
		testLoadSecret(t, secret)
		assert.Equal(t, expSecret, os.Getenv(varname), "sets env var")
		assert.Equal(t, "shhh\n", readTestFile(t, secretLog), "secret not called again (cached)")

		secret.passphrase = "invalidate!"

		os.Unsetenv(varname)
		testLoadSecret(t, secret)
		assert.Equal(t, expSecret, os.Getenv(varname), "sets env var")
		assert.Equal(t, "shhh\nagain\n", readTestFile(t, secretLog), "secret called again (invalid cache)")

		os.Unsetenv(varname)
		testLoadSecret(t, secret)
		assert.Equal(t, expSecret, os.Getenv(varname), "sets env var")
		assert.Equal(t, "shhh\nagain\n", readTestFile(t, secretLog), "secret cached")

		appendToTestFile(t, secretCacheFile, "x")
		os.Unsetenv(varname)
		testLoadSecret(t, secret)
		assert.Equal(t, expSecret, os.Getenv(varname), "sets env var")
		assert.Equal(t, "shhh\nagain\nagain\n", readTestFile(t, secretLog), "cache corrupted")

		os.Setenv(logvarname, "still")

		ioutil.WriteFile(secretCacheFile, []byte("x"), 0600)
		os.Unsetenv(varname)
		testLoadSecret(t, secret)
		assert.Equal(t, expSecret, os.Getenv(varname), "sets env var")
		assert.Equal(t, "shhh\nagain\nagain\nstill\n", readTestFile(t, secretLog), "cache corrupted")

		os.Unsetenv(varname)
		testLoadSecret(t, secret)
		assert.Equal(t, expSecret, os.Getenv(varname), "sets env var")
		assert.Equal(t, "shhh\nagain\nagain\nstill\n", readTestFile(t, secretLog), "cached again")
	})

	t.Run("errors", func(t *testing.T) {
		os.Unsetenv("MUSS_TEST_PASSPHRASE")

		cfg := map[string]interface{}{
			"secrets": map[string]interface{}{
				"commands": map[string]interface{}{
					"some": map[string]interface{}{
						"exec": []string{secretCmdPath, "something"},
					},
				},
			},
		}

		varname := "MUSS_TEST_SECRET_VAR"

		secretSpec := map[string]interface{}{
			"some":    "string",
			"varname": varname,
		}

		assert.Equal(t,
			"value for secret args must be a list",
			testSecretError(t, cfg, secretSpec))
		secretSpec["some"] = []string{"list"}

		assert.Equal(t,
			"a passphrase is required to use secrets",
			testSecretError(t, cfg, secretSpec))

		subMap(cfg, "secrets")["passphrase"] = "static"

		assert.Equal(t,
			"passphrase should contain a variable so it isn't plain text",
			testSecretError(t, cfg, secretSpec))

		os.Unsetenv("MUSS_TEST_PASSPHRASE")
		subMap(cfg, "secrets")["passphrase"] = "$MUSS_TEST_PASSPHRASE"

		assert.Equal(t,
			"a passphrase is required to use secrets",
			testSecretError(t, cfg, secretSpec))

		os.Setenv("MUSS_TEST_PASSPHRASE", "foo")
		secretSpec["exec"] = []string{"echo", "nerts"}

		assert.Regexp(t,
			`secret cannot have multiple commands: ("some" and "exec"|"exec" and "some")`,
			testSecretError(t, cfg, secretSpec))
	})
}

func testSecretError(t *testing.T, cfg ProjectConfig, spec map[string]interface{}) string {
	t.Helper()
	_, err := parseSecret(cfg, spec)
	if err == nil {
		t.Fatal("expected err, got nil")
	}
	return err.Error()
}

func testLoadSecret(t *testing.T, secret *secretCmd) {
	t.Helper()
	if err := secret.load(); err != nil {
		t.Fatalf("failed to load secret: %s", err)
	}
}

func appendToTestFile(t *testing.T, file, suffix string) {
	t.Helper()
	f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.Write([]byte(suffix)); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}
