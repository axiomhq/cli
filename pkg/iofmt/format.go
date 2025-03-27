package iofmt

import "fmt"

//go:generate go tool stringer -type=Format -linecomment -output=format_string.go

// Format is an output format.
type Format uint8

const (
	// Table formats output in tabular style.
	Table Format = iota + 1 // table
	// JSON formats output as one or more JSON objects.
	JSON // json
)

// Formats returns all supported formats.
func Formats() []Format {
	return []Format{Table, JSON}
}

// FormatFromString parses a supported Format from its string representation.
func FormatFromString(s string) (Format, error) {
	var format Format

	switch s {
	case Table.String():
		format = Table
	case JSON.String():
		format = JSON
	default:
		return 0, fmt.Errorf("unknown format %q", s)
	}

	return format, nil
}
