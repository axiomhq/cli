package utils

import (
	"fmt"
	"strconv"

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
