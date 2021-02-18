package client

import (
	"crypto/tls"
	"net/http"

	"github.com/axiomhq/axiom-go/axiom"

	"github.com/axiomhq/cli/pkg/version"
)

// New returns a new Axiom client.
func New(baseURL, accessToken, orgID string, insecure bool, options ...axiom.Option) (*axiom.Client, error) {
	httpClient := axiom.DefaultHTTPClient()

	if insecure {
		if transport, ok := httpClient.Transport.(*http.Transport); ok {
			// nolint:gosec
			transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		}
	}

	options = append(options, axiom.SetUserAgent("axiom-cli/"+version.Release()))
	options = append(options, axiom.SetBaseURL(baseURL))
	options = append(options, axiom.SetClient(httpClient))
	return axiom.NewCloudClient(accessToken, orgID, options...)
}
