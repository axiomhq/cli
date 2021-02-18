package surveyext

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"
)

// ValidateURL validates that the given input is a valid url.
func ValidateURL(val interface{}) error {
	rawurl, ok := val.(string)
	if !ok {
		return fmt.Errorf("url cannot be of type %v", reflect.TypeOf(val).Name())
	}

	if !strings.HasPrefix(rawurl, "http://") && !strings.HasPrefix(rawurl, "https://") {
		rawurl = "https://" + rawurl
	}

	if _, err := url.ParseRequestURI(rawurl); err != nil {
		return err
	}

	return nil
}
