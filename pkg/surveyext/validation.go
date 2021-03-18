package surveyext

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/axiomhq/cli/internal/client"
)

// ValidateURL validates that the given input is a valid url.
func ValidateURL(val interface{}) error {
	rawURL, ok := val.(string)
	if !ok {
		return fmt.Errorf("url cannot be of type %v", reflect.TypeOf(val).Name())
	}

	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "https://" + rawURL
	}

	if _, err := url.ParseRequestURI(rawURL); err != nil {
		return err
	}

	return nil
}

// ValidateToken validates that the given input is a valid Axiom access token
// (personal or ingest).
func ValidateToken(val interface{}) error {
	token, ok := val.(string)
	if !ok {
		return fmt.Errorf("token cannot be of type %v", reflect.TypeOf(val).Name())
	}

	if !client.IsPersonalToken(token) && !client.IsIngestToken(token) {
		return errors.New("token is not an axiom access token (missing prefix)")
	}

	return nil
}
