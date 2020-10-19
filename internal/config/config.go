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

// All available token types.
const (
	Ingest   = "ingest"
	Personal = "personal"
)

func defaultConfigFile() string {
	dir, _ := homedir.Dir()
	return path.Join(dir, ".axiom.toml")
}

var defaultConfig = Config{
	Deployments: make(map[string]Deployment),

	ConfigFilePath: defaultConfigFile(),

	tree: &toml.Tree{},
}

// Config is the global Axiom CLI configuration.
type Config struct {
	ActiveDeployment string                `toml:"active_deployment" envconfig:"deployment"`
	Deployments      map[string]Deployment `toml:"deployments"`

	URLOverride   string `toml:"-" envconfig:"url"`
	TokenOverride string `toml:"-" envconfig:"token"`

	ConfigFilePath string `toml:"-"`

	tree *toml.Tree
}

// GetActiveDeployment returns the configured deployment with overrides applied,
// if given.
func (c *Config) GetActiveDeployment() (Deployment, bool) {
	dep, ok := c.Deployments[c.ActiveDeployment]
	if !ok {
		if c.URLOverride != "" || c.TokenOverride != "" {
			dep = Deployment{
				URL:   c.URLOverride,
				Token: c.TokenOverride,
			}
			return dep, true
		}
		return Deployment{}, false
	}

	if c.URLOverride != "" {
		dep.URL = c.URLOverride
	}
	if c.TokenOverride != "" {
		dep.Token = c.TokenOverride
	}

	return dep, true
}

// Deployment is the configuration for an Axiom instance.
type Deployment struct {
	URL       string `toml:"url"`
	Token     string `toml:"token"`
	TokenType string `toml:"token_type"`
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

	// If only one deployment is configured, make it the active one.
	if len(config.Deployments) == 1 {
		for k := range config.Deployments {
			config.ActiveDeployment = k
			break
		}
	}

	return config, nil
}

// DeploymentAliases returns a sorted slice of deployment aliases.
func (c *Config) DeploymentAliases() []string {
	res := make([]string, 0, len(c.Deployments))
	for k := range c.Deployments {
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
	res := make([]string, 0, len(c.Deployments)*2+1) // 2 fields for each deployment plus the "active_deployment" one
	res = append(res, "active_deployment")
	for k := range c.Deployments {
		base := strings.Join([]string{"deployments", k}, ".")
		res = append(res, strings.Join([]string{base, "url"}, "."))
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
