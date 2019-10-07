package config

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha512"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"os/exec"
	"path"

	"golang.org/x/crypto/nacl/secretbox"
	"golang.org/x/crypto/pbkdf2"
)

var envFileDir string
var secretDir string

type secretCmd struct {
	varname    string
	exec       []string
	passphrase string
}

var projectSecrets []*secretCmd

func init() {
	findCacheRoot()
}

func findCacheRoot() {
	cache, cacheErr := os.UserCacheDir()
	if cacheErr != nil {
		panic(cacheErr)
	}

	setCacheRoot(cache)
}

func setCacheRoot(dir string) {
	wd, wdErr := os.Getwd()
	if wdErr != nil {
		panic(wdErr)
	}

	secretDir = path.Join(dir, ".muss", genFileName(path.Clean(wd)), "secrets")
}

func parseSecret(cfg ProjectConfig, spec map[string]interface{}) (result *secretCmd, err error) {
	var cmd string
	var args []string
	var varname string

	for k, v := range spec {
		switch k {
		case "varname":
			varname = v.(string)
		default:
			if cmd != "" {
				err = fmt.Errorf("secret cannot have multiple commands: %q and %q", cmd, k)
				return
			}
			cmd = k
			var ok bool
			args, ok = stringSlice(v)
			if !ok {
				err = fmt.Errorf("value for secret args must be a list")
				return
			}
		}
	}

	secretConfig := subMap(cfg, "secrets")
	cmdargs, cErr := newSecretCommand(secretConfig, cmd, args)
	if cErr != nil {
		err = cErr
		return
	}

	passphrase, _ := secretConfig["passphrase"].(string)
	var expandedPassphrase string
	if passphrase != "" {
		expandedPassphrase = expandWarnOnEmpty(passphrase)
		if passphrase == expandedPassphrase {
			err = fmt.Errorf("passphrase should contain a variable so it isn't plain text")
			return
		}
	}
	if expandedPassphrase == "" {
		err = fmt.Errorf("a passphrase is required to use secrets")
		return
	}

	return &secretCmd{
		varname:    varname,
		exec:       cmdargs,
		passphrase: expandedPassphrase,
	}, nil
}

func (s *secretCmd) get() (string, error) {
	var content []byte

	// See if we already have the secret cached.
	cacheFile := path.Join(secretDir, genFileName(s.exec))
	if fileContent, err := ioutil.ReadFile(cacheFile); err == nil {
		content = s.decrypt(fileContent)
	}

	// If we don't have a cached value, run the command.
	if len(content) == 0 {
		var stdout, stderr bytes.Buffer
		cmd := exec.Command(s.exec[0], s.exec[1:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("failed to get secret: %s", stderr.String())
		}
		content = bytes.TrimRight(stdout.Bytes(), "\n")

		// Cache it for next time.
		encrypted := s.encrypt(content)
		if len(encrypted) > 0 {
			writePrivateFile(cacheFile, encrypted)
		}
	}

	return string(content), nil
}

func (s *secretCmd) load() error {
	// For a single value...
	if s.varname != "" {
		// Only get it if not already set.
		if _, ok := os.LookupEnv(s.varname); !ok {
			val, err := s.get()
			if err != nil {
				return err
			}
			os.Setenv(s.varname, val)
		}
	} else {
		return fmt.Errorf("non-varname secrets not implemented yet")
	}
	return nil
}

const (
	secretIterationsLen = 3
	secretNonceLen      = 24
	secretNonceStart    = secretIterationsLen
	secretNonceEnd      = secretIterationsLen + secretNonceLen
	secretSaltLen       = 32
	secretSaltStart     = secretNonceEnd
	secretSaltEnd       = secretSaltStart + secretSaltLen
	secretKeyLen        = 32
	secretKeyStart      = secretSaltEnd
	secretKeyEnd        = secretKeyStart + secretKeyLen
	secretPrefixLen     = secretKeyEnd
)

func (s *secretCmd) encrypt(content []byte) []byte {
	nonce := [secretNonceLen]byte{}
	if _, err := rand.Read(nonce[:]); err != nil {
		return nil
	}

	// vaulted uses 17 and 18, but we will lower it for speed:
	// $ muss wrap env
	// with 3 secrets (sequentially):
	// 17,18 -> .85s
	// 16,17 -> .53s
	// 15,16 -> .29s
	// 14,15 -> .15s (feels pretty snappy)
	// 13,14 -> .10s
	// no secrets: .02s (:sofantastic:)
	iterations := 1 << 14
	if r, err := rand.Int(rand.Reader, big.NewInt(1<<15)); err == nil {
		iterations += int(r.Int64())
	}

	salt := [secretSaltLen]byte{}
	if _, err := rand.Read(salt[:]); err != nil {
		return nil
	}

	key := [secretKeyLen]byte{}
	derivedKey := pbkdf2.Key([]byte(s.passphrase), salt[:], iterations, secretKeyLen, sha512.New)
	copy(key[:], derivedKey[:])

	sealed := secretbox.Seal(nil, content, &nonce, &key)

	result := make([]byte, secretPrefixLen+len(sealed))
	copy(result[:], []byte{byte(iterations >> 16), byte(iterations & 0xffff >> 8), byte(iterations & 0xff)})
	copy(result[secretNonceStart:secretNonceEnd], nonce[:])
	copy(result[secretSaltStart:secretSaltEnd], salt[:])
	copy(result[secretKeyStart:secretKeyEnd], key[:])
	copy(result[secretPrefixLen:], sealed[:])

	return result
}

func (s *secretCmd) decrypt(content []byte) (result []byte) {
	// Don't error on slice indexing.
	if len(content) <= secretPrefixLen {
		return nil
	}

	iterations := int(content[0])<<16 + int(content[1])<<8 + int(content[2])

	nonce := [secretNonceLen]byte{}
	copy(nonce[:], content[secretNonceStart:secretNonceEnd])

	salt := [secretSaltLen]byte{}
	copy(salt[:], content[secretSaltStart:secretSaltEnd])

	key := [secretKeyLen]byte{}
	derivedKey := pbkdf2.Key([]byte(s.passphrase), salt[:], iterations, secretKeyLen, sha512.New)
	copy(key[:], derivedKey[:])

	plain, ok := secretbox.Open(nil, content[secretPrefixLen:], &nonce, &key)
	if !ok {
		return nil
	}
	return plain
}

func newSecretCommand(secretConfig map[string]interface{}, name string, args []string) ([]string, error) {
	cmdargs := make([]string, 0)

	// Static command that just runs its args.
	if name == "exec" {
		cmdargs = args
	} else {
		// See if the project configures an alias to simplify service defs.
		if commands, ok := secretConfig["commands"].(map[string]interface{}); ok {
			if command, ok := commands[name].(map[string]interface{}); ok {
				if preArgs, ok := stringSlice(command["exec"]); ok {
					cmdargs = append(preArgs, args...)
				}
			}
		}
	}

	if len(cmdargs) == 0 {
		return nil, fmt.Errorf("failed to prepare secret command '%s'", name)
	}

	return cmdargs, nil
}

func genFileName(args ...interface{}) string {
	h := sha1.New()
	h.Write([]byte(fmt.Sprintf("%#v", args)))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func writePrivateFile(file string, bytes []byte) error {
	if err := os.MkdirAll(path.Dir(file), 0700); err != nil {
		return err
	}
	return ioutil.WriteFile(file, bytes, 0600)
}
