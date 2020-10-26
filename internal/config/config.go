package config

import (
	"io"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/kelseyhightower/envconfig"
	"github.com/mitchellh/go-homedir"
	"github.com/pelletier/go-toml"
)

// Name is the default configuration filename.
const Name = ".axiom.toml"

// Default is the default configuration.
var Default = Config{
	Backends: make(map[string]Backend),
}

// Config is the global Axiom CLI configuration.
type Config struct {
	ActiveBackend string             `toml:"active_backend" envconfig:"backend"`
	Backends      map[string]Backend `toml:"backends"`

	configFilePath string
}

// Backend is the configuration for an Axiom instance.
type Backend struct {
	URL      string `toml:"url"`
	Username string `toml:"username"`
	Token    string `toml:"token"`

	tree *toml.Tree
}

// LoadDefaultConfigFile tries to load the default configuration. It doesn't
// fail if the configuration file cannot be found. However, it fails when the
// found config fails to load for different reasons. On error, the default
// configuration is returned.
func LoadDefaultConfigFile() (Config, error) {
	dir, err := homedir.Dir()
	if err != nil {
		return Default, err
	}

	configFilePath := path.Join(dir, ".axiom.toml")

	config, err := Load(configFilePath)
	if err != nil && !os.IsNotExist(err) {
		return Default, err
	}
	config.configFilePath = configFilePath

	return config, nil
}

// Load loads configuration from a file. Configuration values loaded from the
// environment overwrite the ones from the TOML configuration. On error, the
// default configuration or the configuration processed till the error is
// returned.
func Load(file string) (Config, error) {
	f, err := os.Open(file)
	if err != nil {
		return LoadFromReader(strings.NewReader(""))
	}
	defer f.Close()

	config, err := LoadFromReader(f)
	config.configFilePath = file

	return config, err
}

// LoadFromReader loads configuration from an io.Reader. Configuration values
// loaded from the environment overwrite the ones from the TOML configuration.
// On error, the default configuration or the configuration processed till the
// error is returned.
func LoadFromReader(r io.Reader) (Config, error) {
	config := Default

	if err := toml.NewDecoder(r).Decode(&config); err != nil {
		return config, err
	} else if err = envconfig.Process("axm", &config); err != nil {
		return config, err
	}

	// If only one backend is configured, make it the active one.
	if len(config.Backends) == 1 {
		for k := range config.Backends {
			config.ActiveBackend = k
			break
		}
	}

	return config, nil
}

// BackendAliases returns a sorted slice of backend aliases.
func (c *Config) BackendAliases() []string {
	res := make([]string, 0, len(c.Backends))
	for k := range c.Backends {
		res = append(res, k)
	}
	sort.Strings(res)
	return res
}

// Write the configuration to its configuration file on disk. If it doesn't
// exist, it is created.
func (c *Config) Write() error {
	if c.configFilePath == "" {
		return nil
	}

	f, err := os.Create(c.configFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	return toml.NewEncoder(f).Encode(c)
}
