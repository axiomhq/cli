package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gofrs/flock"
	"github.com/pelletier/go-toml"
)

// AddDeployment saves a new deployment without applying environment overrides
// or changing existing entries. A failed write leaves the old file intact.
func AddDeployment(filename, alias string, deployment Deployment) error {
	// Keep symlinked configs working when replacing the file.
	if _, err := os.Lstat(filename); err == nil {
		filename, err = filepath.EvalSymlinks(filename)
		if err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	// Lock a stable sidecar: renaming the config replaces its inode. Keep the
	// lock file so concurrent creators always lock the same inode.
	lock := flock.New(filename + ".lock")
	if err := lock.Lock(); err != nil {
		return err
	}
	defer lock.Close()

	tree, err := toml.LoadFile(filename)
	if os.IsNotExist(err) {
		tree, err = toml.Load("")
	}
	if err != nil {
		return err
	}
	var stored Config
	if err := tree.Unmarshal(&stored); err != nil {
		return err
	}
	if _, exists := stored.Deployments[alias]; exists {
		return fmt.Errorf("deployment %q already exists", alias)
	}

	// A single deployment is implicitly active when loaded. Keep it active
	// when adding a second deployment, even if no default was written before.
	if stored.ActiveDeployment == "" && len(stored.Deployments) == 1 {
		for existing := range stored.Deployments {
			tree.Set("active_deployment", existing)
		}
	}
	if stored.ActiveDeployment == "" && len(stored.Deployments) == 0 {
		tree.Set("active_deployment", alias)
	}
	data, err := toml.Marshal(deployment)
	if err != nil {
		return err
	}
	entry, err := toml.LoadBytes(data)
	if err != nil {
		return err
	}
	tree.SetPath([]string{"deployments", alias}, entry)

	f, err := os.CreateTemp(filepath.Dir(filename), ".axiom-config-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	defer f.Close()
	if _, err := tree.WriteTo(f); err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), filename)
}
