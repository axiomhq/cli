package client

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/klauspost/compress/gzhttp"

	"github.com/axiomhq/pkg/version"
)

// New returns a new Axiom client.
func New(ctx context.Context, baseURL, accessToken, orgID string, insecure bool) (*axiom.Client, error) {
	if baseURL != "" && !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "https://" + baseURL
	}

	httpTransport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 5 * time.Second,
		ForceAttemptHTTP2:   true,
	}

	if insecure {
		httpTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // This is intended behaviour.
	}

	httpClient := &http.Client{
		Transport: gzhttp.Transport(httpTransport),
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
		options = append(options, axiom.SetToken(accessToken))
	}
	if orgID != "" {
		options = append(options, axiom.SetOrganizationID(orgID))
	}

	client, err := axiom.NewClient(options...)
	if err != nil {
		return nil, err
	}

	return client, client.ValidateCredentials(ctx)
}
