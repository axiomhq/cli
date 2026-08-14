package utils

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/axiomhq/cli/pkg/terminal"
)

// Pluralize a string, if n is greater than one. It uses the given color scheme
// to print the number bold.
func Pluralize(cs *terminal.ColorScheme, s string, n int) string {
	nStr := cs.Bold(strconv.Itoa(n))
	if n == 1 {
		return fmt.Sprintf("%s %s", nStr, s)
	}
	return fmt.Sprintf("%s %ss", nStr, s)
}

// String automatically detects the maximum indentation shared by all lines and
// trims them accordingly.
// Ref: https://github.com/muesli/reflow/blob/v0.2.0/dedent/dedent.go
func Dedent(s string) string {
	lines := strings.Split(s, "\n")
	minIndent := -1

	for _, l := range lines {
		if len(l) == 0 {
			continue
		}

		indent := len(l) - len(strings.TrimLeft(l, " "))
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}

	if minIndent <= 0 {
		return s
	}

	var buf strings.Builder
	for _, l := range lines {
		l = strings.TrimPrefix(l, strings.Repeat(" ", minIndent))
		buf.WriteString(l + "\n")
	}
	return strings.TrimSuffix(buf.String(), "\n")
}
