package cmdutil

import (
	"context"
	"errors"

	"github.com/axiomhq/axiom-go/axiom"

	"github.com/axiomhq/cli/internal/client"
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
func (f *Factory) Client(ctx context.Context) (*axiom.Client, error) {
	deployment, ok := f.Config.GetActiveDeployment()
	if !ok {
		return nil, errors.New("no active deployment set")
	}
	return client.New(ctx, deployment.URL, deployment.Token, deployment.OrganizationID,
		deployment.EdgeURL, deployment.EdgeRegion, f.Config.Insecure)
}
