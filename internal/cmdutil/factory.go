package cmdutil

import (
	"errors"

	"axicode.axiom.co/watchmakers/axiomdb/client"

	"github.com/axiomhq/cli/internal/config"
	"github.com/axiomhq/cli/pkg/terminal"
)

// Factory bundles resources used by most commands.
type Factory struct {
	// Config is the configuration.
	Config *config.Config
	// IO is the IO to be used instead of StdIn, StdOut and StdErr.
	IO *terminal.IO
}

// NewFactory creates a new Factory.
func NewFactory() *Factory {
	return &Factory{
		IO: terminal.NewIO(),
	}
}

// Client returns an Axiom client configured to talk to the instance specified
// by the configuration.
func (f *Factory) Client() (*client.Client, error) {
	backend, ok := f.Config.Backends[f.Config.ActiveBackend]
	if !ok {
		return nil, errors.New("no active backend set")
	}
	return client.NewClient(backend.URL)
}
