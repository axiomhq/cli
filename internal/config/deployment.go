package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gofrs/flock"
	"github.com/pelletier/go-toml"
)

// AddDeployment saves a new deployment without applying environment overrides or changing existing entries.
// A failed write leaves the old file intact.
func AddDeployment(ctx context.Context, filename, alias string, deployment Deployment) error {
	// NOTE: replace the symlink target so existing config links keep working.
	if _, err := os.Lstat(filename); err == nil {
		filename, err = filepath.EvalSymlinks(filename)
		if err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	// NOTE: renaming the config changes its inode, so concurrent saves lock a stable sidecar file.
	// Keep the lock file so later saves use the same lock.
	lock := flock.New(filename + ".lock")
	defer lock.Close()
	if _, err := lock.TryLockContext(ctx, 100*time.Millisecond); err != nil {
		return err
	}

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

	// NOTE: loading a single deployment makes it active even when no default is stored.
	// Preserve that choice when adding a second deployment.
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
