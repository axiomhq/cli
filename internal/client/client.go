package client

import (
	"crypto/tls"
	"errors"
	"net/http"
	"strings"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/pkg/version"
)

// New returns a new Axiom client.
func New(baseURL, accessToken, orgID string, insecure bool, options ...axiom.Option) (*axiom.Client, error) {
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "https://" + baseURL
	}

	httpClient := axiom.DefaultHTTPClient()

	if insecure {
		transport, ok := httpClient.Transport.(*http.Transport)
		if !ok {
			return nil, errors.New("could not set insecure mode")
		}
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec
	}

	options = append(options, axiom.SetUserAgent("axiom-cli/"+version.Release()))
	options = append(options, axiom.SetBaseURL(baseURL))
	options = append(options, axiom.SetClient(httpClient))
	return axiom.NewCloudClient(accessToken, orgID, options...)
}

// IsPersonalToken returns true if the given token is a personal token.
func IsPersonalToken(token string) bool {
	return strings.HasPrefix(token, "xapt-")
}

// IsIngestToken returns true if the given token is an ingest token.
func IsIngestToken(token string) bool {
	return strings.HasPrefix(token, "xait-")
}
