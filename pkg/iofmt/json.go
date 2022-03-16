package iofmt

import (
	"encoding/json"
	"io"

	"github.com/nwidger/jsoncolor"
)

// jsonEncoder is an interface implemented by most JSON encoders.
type jsonEncoder interface {
	Encode(any) error
}

// FormatToJSON formats the given data in JSON format.
func FormatToJSON(w io.Writer, v any, colorEnabled bool) error {
	var encoder jsonEncoder
	if colorEnabled {
		encoder = jsoncolor.NewEncoder(w)
	} else {
		encoder = json.NewEncoder(w)
	}

	return encoder.Encode(v)
}
