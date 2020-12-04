package cmdutil

import (
	"errors"

	"github.com/axiomhq/axiom-go/axiom"

	axiomClient "github.com/axiomhq/cli/internal/client"
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

// Client returns an Axiom client configured to talk to the active deployment.
func (f *Factory) Client() (*axiom.Client, error) {
	deployment, ok := f.Config.GetActiveDeployment()
	if !ok {
		return nil, errors.New("no active deployment set")
	}
	return axiomClient.New(deployment.URL, deployment.Token)
}
