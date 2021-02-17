package client

import (
	"github.com/axiomhq/axiom-go/axiom"

	"github.com/axiomhq/cli/pkg/version"
)

// New returns a new Axiom client.
func New(baseURL, accessToken, orgID string, options ...axiom.Option) (*axiom.Client, error) {
	options = append(options, axiom.SetUserAgent("axiom-cli/"+version.Release()))
	options = append(options, axiom.SetBaseURL(baseURL))
	return axiom.NewCloudClient(accessToken, orgID, options...)
}
