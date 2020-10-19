package config

import (
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/kelseyhightower/envconfig"
	"github.com/mitchellh/go-homedir"
	"github.com/pelletier/go-toml"
)

func defaultConfigFile() string {
	dir, _ := homedir.Dir()
	return path.Join(dir, ".axiom.toml")
}

var defaultConfig = Config{
	Backends: make(map[string]Backend),

	ConfigFilePath: defaultConfigFile(),
	tree:           &toml.Tree{},
}

// Config is the global Axiom CLI configuration.
type Config struct {
	ActiveBackend string             `toml:"active_backend" envconfig:"backend"`
	Backends      map[string]Backend `toml:"backends"`

	ConfigFilePath string `toml:"-"`

	tree *toml.Tree
}

// Backend is the configuration for an Axiom instance.
type Backend struct {
	URL      string `toml:"url"`
	Username string `toml:"username"`
	Token    string `toml:"token"`
}

// LoadDefault tries to load the default configuration. It behaves like Load()
// but doesn't fail if the configuration file doesn't exist.
func LoadDefault() (*Config, error) {
	config, err := Load(defaultConfigFile())
	if err != nil && !os.IsNotExist(err) {
		return config, err
	}

	return config, nil
}

// Load the configuration. It behaves like LoadFromReader() but opens a file
// for reading the TOML configuration.
func Load(configFilePath string) (*Config, error) {
	f, err := os.Open(configFilePath)
	if err != nil {
		// Make sure the environment variables are processed while persisting
		// the open file error.
		config, _ := LoadFromReader(strings.NewReader(""))
		return config, err
	}
	defer f.Close()

	config, err := LoadFromReader(f)
	config.ConfigFilePath = configFilePath

	return config, err
}

// LoadFromReader loads configuration from an io.Reader. Configuration values
// loaded from the environment overwrite the ones from the TOML configuration.
// On error, the default configuration or the configuration processed till the
// error occurred is returned.
func LoadFromReader(r io.Reader) (config *Config, err error) {
	config = &defaultConfig

	if config.tree, err = toml.LoadReader(r); err != nil {
		return config, err
	} else if err = config.tree.Unmarshal(config); err != nil {
		return config, err
	} else if err = envconfig.Process("axm", config); err != nil {
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

// Get the string value for the given key.
func (c *Config) Get(key string) (string, error) {
	if !c.tree.Has(key) {
		return "", fmt.Errorf("got no key %q", key)
	}

	val, ok := c.tree.Get(key).(string)
	if !ok {
		return "", fmt.Errorf("value at key %q is not a string", key)
	}
	val = strings.TrimSpace(val)

	if val == "" {
		return "", fmt.Errorf("value at key %q is present but an empty string", key)
	}

	return val, nil
}

// Set the string value at the given key. Existing values are overwritten. The
// key must exist in the configuration.
func (c *Config) Set(key, value string) error {
	if !c.tree.Has(key) {
		return fmt.Errorf("got no key %q", key)
	}

	c.tree.Set(key, value)

	return c.tree.Unmarshal(c)
}

// Keys which are valid arguments to Get() and Set().
func (c *Config) Keys() []string {
	res := make([]string, 0, len(c.Backends)*4+1) // 3 fields for each backend plus the "active_backend" one
	res = append(res, "active_backend")
	for k := range c.Backends {
		base := strings.Join([]string{"backends", k}, ".")
		res = append(res, strings.Join([]string{base, "url"}, "."))
		res = append(res, strings.Join([]string{base, "username"}, "."))
		res = append(res, strings.Join([]string{base, "token"}, "."))
	}
	sort.Strings(res)
	return res
}

// Write the configuration to its configuration file on disk. If it doesn't
// exist, it is created.
func (c *Config) Write() error {
	f, err := os.Create(c.ConfigFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	if err = toml.NewEncoder(f).Order(toml.OrderPreserve).Encode(c); err != nil {
		return err
	}

	return f.Sync()
}
