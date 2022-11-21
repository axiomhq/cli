package client

import (
	"net/url"
	"strings"
)

// CloudURL is the Axiom Cloud URL.
const CloudURL = "https://cloud.axiom.co"

// IsCloudURL returns true if the given URL is an Axiom Cloud URL.
func IsCloudURL(s string) bool {
	if s != "" && !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
		s = "https://" + s
	}

	u, err := url.ParseRequestURI(s)
	if err != nil {
		return false
	}

	cu, err := url.ParseRequestURI(CloudURL)
	if err != nil {
		return false
	}

	return u.Host == cu.Host
}

// IsPersonalToken returns true if the given token is a personal token.
func IsPersonalToken(token string) bool {
	return strings.HasPrefix(token, "xapt-")
}
