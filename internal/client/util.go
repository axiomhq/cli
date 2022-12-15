package client

import (
	"fmt"
	"net/url"
	"strings"
)

// URLs of the hosted Axiom deployment.
const (
	BaseURL = "https://axiom.co"

	AppURL   = "https://app.axiom.co"
	APIURL   = "https://api.axiom.co"
	LoginURL = "https://login.axiom.co"
)

// IsPersonalToken returns true if the given token is a personal token.
func IsPersonalToken(token string) bool {
	return strings.HasPrefix(token, "xapt-")
}

// GetAppURL returns the app URL for the given deployment URL.
func GetAppURL(baseURL string) (string, error) {
	return getURL(baseURL, "app", AppURL)
}

// GetAPIURL returns the api URL for the given deployment URL.
func GetAPIURL(baseURL string) (string, error) {
	return getURL(baseURL, "api", APIURL)
}

// GetLoginURL returns the login URL for the given deployment URL.
func GetLoginURL(baseURL string) (string, error) {
	return getURL(baseURL, "login", LoginURL)
}

func getURL(baseURL, subdomainToAdd, ignoreCase string) (string, error) {
	// Make sure to accept urls without a scheme.
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "https://" + baseURL
	}

	u, err := url.ParseRequestURI(baseURL)
	if err != nil {
		return "", err
	}

	// If this is the ignore case, return it.
	if s := u.String(); s == ignoreCase {
		return s, nil
	}

	// Prepend the subdomain.
	u.Host = fmt.Sprintf("%s.%s", subdomainToAdd, u.Host)

	// Make sure it is valid.
	if u, err = url.ParseRequestURI(u.String()); err != nil {
		return "", err
	}

	return u.String(), nil
}
