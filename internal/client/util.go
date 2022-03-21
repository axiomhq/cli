package client

import (
	"net/url"
	"strings"

	"github.com/axiomhq/axiom-go/axiom"
)

// IsCloudURL returns true if the given URL is an Axiom Cloud URL.
func IsCloudURL(s string) bool {
	if s != "" && !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
		s = "https://" + s
	}

	u, err := url.ParseRequestURI(s)
	if err != nil {
		return false
	}

	cu, err := url.ParseRequestURI(axiom.CloudURL)
	if err != nil {
		return false
	}

	return u.Host == cu.Host
}
