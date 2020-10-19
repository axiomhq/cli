package doc_test

import (
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/stretchr/testify/assert"

	"github.com/axiomhq/cli/pkg/doc"
)

func Test_String(t *testing.T) {
	input := `
		NAME: Set to any value to avoid printing ANSI escape sequences for color
		output.
		
		CLICOLOR: Set to "0" to disable printing ANSI colors in output.
		
		CLICOLOR_FORCE: Set to a value other than "0" to keep ANSI colors in
		output even when the output is piped.
	`

	want := heredoc.Doc(`
		NAME: Set to any value to avoid printing ANSI escape sequences for color output.
		
		CLICOLOR: Set to "0" to disable printing ANSI colors in output.
		
		CLICOLOR_FORCE: Set to a value other than "0" to keep ANSI colors in output even when the output is piped.
	`)

	assert.Equal(t, want, doc.String(input))
}
