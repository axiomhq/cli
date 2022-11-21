package surveyext

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/axiomhq/cli/internal/client"

	"github.com/AlecAivazis/survey/v2"
)

// ValidateURL validates that the given input is a valid url.
func ValidateURL(val any) error {
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
func ValidateToken(val any) error {
	if token, ok := val.(string); !ok {
		return fmt.Errorf("token cannot be of type %v", reflect.TypeOf(val).Name())
	} else if !client.IsPersonalToken(token) {
		return errors.New("token is not a personal access token (missing 'xapt-' prefix)")
	}
	return nil
}

// NotIn returns a validation function that returns an error if the value is in
// the given list.
func NotIn(ss []string) survey.Validator {
	return func(val any) error {
		v, ok := val.(string)
		if !ok {
			return fmt.Errorf("input cannot be of type %v", reflect.TypeOf(val).Name())
		}

		for _, s := range ss {
			if v == s {
				return fmt.Errorf("input cannot be %q", v)
			}
		}

		return nil
	}
}
