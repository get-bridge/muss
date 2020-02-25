package config

import (
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha512"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"path"
	"sync"

	"golang.org/x/crypto/nacl/secretbox"
	"golang.org/x/crypto/pbkdf2"
)

var secretDir string

type secretCmd struct {
	name string
	*envCmd
	passphrase string
}

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

type secretSetup struct {
	done    bool
	mutex   sync.Mutex
	envCmds []envLoader
}

var secretEnvCommands = make(map[string]*secretSetup)

func parseSecret(cfg *ProjectConfig, spec map[string]interface{}) (*secretCmd, error) {
	var name string
	var args []string
	var varname string
	var parse bool

	for k, v := range spec {
		switch k {
		case "varname":
			varname = v.(string)
		case "parse":
			parse = v.(bool)
		default:
			if name != "" {
				return nil, fmt.Errorf("secret cannot have multiple commands: %q and %q", name, k)
			}
			name = k
			var ok bool
			args, ok = stringSlice(v)
			if !ok {
				return nil, fmt.Errorf("value for secret args must be a list")
			}
		}
	}

	cmdargs := make([]string, 0)

	// Default to global.
	passphrase := cfg.SecretPassphrase

	// Static command that just runs its args.
	if name == "exec" {
		cmdargs = args
	} else {
		// See if the project configures an alias to simplify service defs.
		if cfg.SecretCommands != nil {
			if command, ok := cfg.SecretCommands[name].(map[string]interface{}); ok {

				if preArgs, ok := stringSlice(command["exec"]); ok {
					cmdargs = append(preArgs, args...)
				}

				if pass, ok := command["passphrase"].(string); ok && pass != "" {
					passphrase = pass
				}

				ecs, envErr := parseEnvCommands(command["env_commands"])
				if envErr != nil {
					return nil, envErr
				}
				secretEnvCommands[name] = &secretSetup{envCmds: ecs}
			}
		}
	}

	if len(cmdargs) == 0 {
		return nil, fmt.Errorf("failed to prepare secret command '%s'", name)
	}

	return &secretCmd{
		name: name,
		envCmd: &envCmd{
			exec:    cmdargs,
			parse:   parse,
			varname: varname,
		},
		passphrase: passphrase,
	}, nil
}

func (s *secretCmd) Passphrase() ([]byte, error) {
	var expandedPassphrase string
	if s.passphrase != "" {
		expandedPassphrase = expandWarnOnEmpty(s.passphrase)
		if s.passphrase == expandedPassphrase {
			return nil, fmt.Errorf("passphrase should contain a variable so it isn't plain text")
		}
	}
	if expandedPassphrase == "" {
		return nil, fmt.Errorf("a passphrase is required to use secrets")
	}
	return []byte(expandedPassphrase), nil
}

func (s *secretCmd) Value() ([]byte, error) {
	if err := runSecretSetup(s.name); err != nil {
		return nil, err
	}

	passphrase, err := s.Passphrase()
	if err != nil {
		return nil, err
	}

	var content []byte

	// See if we already have the secret cached.
	cacheFile := path.Join(secretDir, genFileName(s.exec))
	if fileContent, err := ioutil.ReadFile(cacheFile); err == nil {
		content = s.decrypt(passphrase, fileContent)
	}

	// If we don't have a cached value, run the command.
	if len(content) == 0 {
		var err error
		content, err = s.envCmd.Value()
		if err != nil {
			return nil, fmt.Errorf("failed to get secret: %s", err)
		}

		// Cache it for next time.
		encrypted := s.encrypt(passphrase, content)
		if len(encrypted) > 0 {
			writePrivateFile(cacheFile, encrypted)
		}
	}

	return content, nil
}

func runSecretSetup(name string) error {
	s := secretEnvCommands[name]
	if s == nil {
		return nil
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.done {
		if err := loadEnvFromCmds(s.envCmds...); err != nil {
			return err
		}
		s.done = true
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

func (s *secretCmd) encrypt(passphrase, content []byte) []byte {
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
	derivedKey := pbkdf2.Key(passphrase, salt[:], iterations, secretKeyLen, sha512.New)
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

func (s *secretCmd) decrypt(passphrase, content []byte) (result []byte) {
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
	derivedKey := pbkdf2.Key(passphrase, salt[:], iterations, secretKeyLen, sha512.New)
	copy(key[:], derivedKey[:])

	plain, ok := secretbox.Open(nil, content[secretPrefixLen:], &nonce, &key)
	if !ok {
		return nil
	}
	return plain
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
