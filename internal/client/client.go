package client

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"strings"

	"github.com/axiomhq/axiom-go/axiom"

	"github.com/axiomhq/pkg/version"
)

// New returns a new Axiom client.
func New(ctx context.Context, baseURL, accessToken, orgID string, insecure bool) (*axiom.Client, error) {
	if baseURL != "" && !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
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

	options := []axiom.Option{
		axiom.SetNoEnv(),
		axiom.SetUserAgent("axiom-cli/" + version.Release()),
		axiom.SetClient(httpClient),
	}

	if baseURL != "" {
		options = append(options, axiom.SetURL(baseURL))
	}
	if accessToken != "" {
		options = append(options, axiom.SetAccessToken(accessToken))
	}
	if orgID != "" {
		options = append(options, axiom.SetOrgID(orgID))
	}

	client, err := axiom.NewClient(options...)
	if err != nil {
		return nil, err
	}

	return client, client.ValidateCredentials(ctx)
}
