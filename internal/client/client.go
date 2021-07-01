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
func New(baseURL, accessToken, orgID string, insecure bool) (*axiom.Client, error) {
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

	return axiom.NewClient(
		axiom.SetURL(baseURL),
		axiom.SetAccessToken(accessToken),
		axiom.SetOrgID(orgID),
		axiom.SetClient(httpClient),
		axiom.SetUserAgent("axiom-cli/"+version.Release()),
	)
}
