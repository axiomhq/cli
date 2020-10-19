package surveyext

import (
	"fmt"
	"net/url"
	"reflect"
)

// ValidateURL validates that the given input is a valid url.
func ValidateURL(val interface{}) error {
	if rawurl, ok := val.(string); !ok {
		return fmt.Errorf("url cannot be of type %v", reflect.TypeOf(val).Name())
	} else if _, err := url.ParseRequestURI(rawurl); err != nil {
		return err
	}
	return nil
}
