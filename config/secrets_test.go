package config

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"

	"gerrit.instructure.com/muss/testutil"
)

func TestSecretCommands(t *testing.T) {
	var secretCmdPath string
	if dir, err := os.Getwd(); err != nil {
		t.Fatalf("failed to get working dir: %s", err)
	} else {
		secretCmdPath = path.Join(dir, "..", "testdata", "bin", "some-secret")
	}

	testutil.WithTempDir(t, func(tmpdir string) {
		findCacheRoot()

		os.Unsetenv("MUSS_TEST_PASSPHRASE")

		cfg := ProjectConfig{
			SecretPassphrase: "$MUSS_TEST_PASSPHRASE",
			SecretCommands: map[string]interface{}{
				"some": map[string]interface{}{
					"exec": []string{secretCmdPath, "something"},
					"env_commands": []interface{}{
						map[string]interface{}{
							"exec":    []string{secretCmdPath, "pre-cmd"},
							"varname": "MUSS_TEST_PASSPHRASE",
						},
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

		secret, err := parseSecret(&cfg, secretSpec)
		if err != nil {
			t.Fatalf("error preparing secret env file: %s", err)
		}

		secretCacheFile := path.Join(secretDir, genFileName([]string{secretCmdPath, "something", "green"}))
		secretLog := "secret-log.txt"

		testutil.NoFileExists(t, secretCacheFile)

		os.Setenv(varname, "oops")
		testLoadSecret(t, secret)
		assert.Equal(t, os.Getenv(varname), "oops", "existing var not overwritten")
		testutil.NoFileExists(t, secretLog) // secret not called

		expSecret := "secret is [something green]"

		os.Unsetenv(varname)
		testLoadSecret(t, secret)
		assert.Equal(t, expSecret, os.Getenv(varname), "sets env var")
		assert.FileExists(t, secretCacheFile)
		assert.Equal(t, "shhh p\nshhh s\n", testutil.ReadFile(t, secretLog), "pre-cmd and secret each called once")

		os.Setenv(logvarname, "again")

		os.Unsetenv(varname)
		testLoadSecret(t, secret)
		assert.Equal(t, expSecret, os.Getenv(varname), "sets env var")
		assert.Equal(t, "shhh p\nshhh s\n", testutil.ReadFile(t, secretLog), "neither called again (cached)")

		os.Setenv("MUSS_TEST_PASSPHRASE", "invalidate!")

		os.Unsetenv(varname)
		testLoadSecret(t, secret)
		assert.Equal(t, expSecret, os.Getenv(varname), "sets env var")
		assert.Equal(t, "shhh p\nshhh s\nagain s\n", testutil.ReadFile(t, secretLog), "secret called again (invalid cache)")

		os.Unsetenv(varname)
		testLoadSecret(t, secret)
		assert.Equal(t, expSecret, os.Getenv(varname), "sets env var")
		assert.Equal(t, "shhh p\nshhh s\nagain s\n", testutil.ReadFile(t, secretLog), "secret cached")

		appendToTestFile(t, secretCacheFile, "x")
		os.Unsetenv(varname)
		testLoadSecret(t, secret)
		assert.Equal(t, expSecret, os.Getenv(varname), "sets env var")
		assert.Equal(t, "shhh p\nshhh s\nagain s\nagain s\n", testutil.ReadFile(t, secretLog), "cache corrupted")

		os.Setenv(logvarname, "still")

		ioutil.WriteFile(secretCacheFile, []byte("x"), 0600)
		os.Unsetenv(varname)
		testLoadSecret(t, secret)
		assert.Equal(t, expSecret, os.Getenv(varname), "sets env var")
		assert.Equal(t, "shhh p\nshhh s\nagain s\nagain s\nstill s\n", testutil.ReadFile(t, secretLog), "cache corrupted")

		os.Unsetenv(varname)
		testLoadSecret(t, secret)
		assert.Equal(t, expSecret, os.Getenv(varname), "sets env var")
		assert.Equal(t, "shhh p\nshhh s\nagain s\nagain s\nstill s\n", testutil.ReadFile(t, secretLog), "cached again")

		t.Run("multiple vars in one command", func(t *testing.T) {
			os.Setenv("MUSS_TEST_PASSPHRASE", "howdy")

			os.Unsetenv("MUSS_TEST_LINE_1_SETUP")
			os.Unsetenv("MUSS_TEST_LINE_2_SETUP")
			os.Unsetenv("MUSS_TEST_LINE_1_SECRET")
			os.Unsetenv("MUSS_TEST_LINE_2_SECRET")

			cfg := &ProjectConfig{
				SecretPassphrase: "$MUSS_TEST_PASSPHRASE",
				SecretCommands: map[string]interface{}{
					"some": map[string]interface{}{
						"exec": []string{secretCmdPath, "--multi"},
						"env_commands": []interface{}{
							map[string]interface{}{
								"exec":  []string{secretCmdPath, "--multi", "SETUP"},
								"parse": true,
							},
						},
					},
				},
			}

			secretSpec := map[string]interface{}{
				"some":  []string{"SECRET"},
				"parse": true,
			}

			secret, err := parseSecret(cfg, secretSpec)
			if err != nil {
				t.Fatalf("error preparing secret env file: %s", err)
			}

			secretLog := "secret-log.txt"

			os.Remove(secretLog)
			testutil.NoFileExists(t, secretLog)

			testLoadSecret(t, secret)

			assert.Equal(t, "foo bar baz", os.Getenv("MUSS_TEST_LINE_1_SETUP"), "set first env var")
			assert.Equal(t, "something", os.Getenv("MUSS_TEST_LINE_2_SETUP"), "set second env var")
			assert.Equal(t, "foo bar baz", os.Getenv("MUSS_TEST_LINE_1_SECRET"), "set first env var")
			assert.Equal(t, "something", os.Getenv("MUSS_TEST_LINE_2_SECRET"), "set second env var")

			assert.Equal(t, "multi SETUP\nmulti SECRET\n", testutil.ReadFile(t, secretLog), "run once")

			testLoadSecret(t, secret)

			// setup only gets called once and the secret gets cached.
			assert.Equal(t, "multi SETUP\nmulti SECRET\n", testutil.ReadFile(t, secretLog), "neither runs again")
		})
	})

	t.Run("errors", func(t *testing.T) {
		os.Unsetenv("MUSS_TEST_PASSPHRASE")

		varname := "MUSS_TEST_SECRET_VAR"
		os.Unsetenv(varname)

		cfg := &ProjectConfig{
			SecretCommands: map[string]interface{}{
				"some": map[string]interface{}{
					"exec": []string{secretCmdPath, "something"},
				},
			},
		}

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

		cfg.SecretPassphrase = "static"

		assert.Equal(t,
			"passphrase should contain a variable so it isn't plain text",
			testSecretError(t, cfg, secretSpec))

		os.Unsetenv("MUSS_TEST_PASSPHRASE")
		cfg.SecretPassphrase = "$MUSS_TEST_PASSPHRASE"

		assert.Equal(t,
			"a passphrase is required to use secrets",
			testSecretError(t, cfg, secretSpec))

		os.Setenv("MUSS_TEST_PASSPHRASE", "foo")
		secretSpec["exec"] = []string{"echo", "nerts"}

		assert.Regexp(t,
			`secret cannot have multiple commands: ("some" and "exec"|"exec" and "some")`,
			testSecretError(t, cfg, secretSpec))

		cfg.SecretCommands["some"] = map[string]interface{}{
			"exec": []string{secretCmdPath, "--no-var"},
		}
		secretSpec = map[string]interface{}{
			"some": []string{},
		}
		assert.Equal(t,
			`env command must have either "parse: true" or a "varname"`,
			testSecretError(t, cfg, secretSpec))

		secretSpec["parse"] = true

		assert.Equal(t,
			`failed to parse name=value line: NO_EQUAL_SIGN`,
			testSecretError(t, cfg, secretSpec))

		secretSpec["varname"] = "foo"

		assert.Equal(t,
			`use "parse: true" or "varname", not both`,
			testSecretError(t, cfg, secretSpec))
	})
}

func testSecretError(t *testing.T, cfg *ProjectConfig, spec map[string]interface{}) string {
	t.Helper()

	s, err := parseSecret(cfg, spec)
	// Some errors don't occur until trying to load it.
	if err == nil {
		err = loadEnvFromCmds(s)
	}
	if err == nil {
		t.Fatal("expected err, got nil")
	}
	return err.Error()
}

func testLoadSecret(t *testing.T, secret *secretCmd) {
	t.Helper()
	if err := loadEnvFromCmds(secret); err != nil {
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
