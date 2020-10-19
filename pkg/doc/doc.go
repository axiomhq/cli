package doc

import (
	"bufio"
	"io"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/muesli/reflow/wordwrap"
)

// String enhances heredoc.Doc() with word unwrapping capabilities. This is
// great for correctly wrapping text when not knowing the wrap width upfront.
func String(s string) string {
	s = heredoc.Doc(s)
	r := strings.NewReader(s)
	return unwrap(r)
}

// Wrap is like Doc() but word wraps at the given width but at max 120
// characters.
func Wrap(s string, wrap int) string {
	if wrap > 120 {
		wrap = 120
	}
	s = String(s)
	return wordwrap.String(s, wrap)
}

func unwrap(r io.Reader) string {
	var (
		br  = bufio.NewReader(r)
		buf strings.Builder
	)
	for {
		// Read till newline.
		part, err := br.ReadString('\n')
		if err != nil {
			break
		}

		// Remove the newline and add that part of the string to the buffer.
		part = strings.TrimSuffix(part, "\n")
		buf.WriteString(part)

		next, err := br.ReadByte()
		if err != nil {
			break
		}

		// If the next character is not a newline character, we figure that the
		// previous newline was just added to wrap text in the editor for
		// readability purposes and the sentence goes on. Thus we write a
		// whitespace character to separate the words instead of a newline and
		// then write the read character and continue.
		if next != '\n' {
			buf.WriteByte(' ')
			buf.WriteByte(next)
			continue
		}

		// If the next character is a newline character, we figure it is in
		// place to logically separate lines of text. They are preserved. Two
		// newlines are written upfront: The previously trimmed one and the just
		// read one.
		buf.WriteString("\n\n")
		for {
			ch, err := br.ReadByte()
			if err != nil {
				break
			}

			buf.WriteByte(ch)

			if ch != '\n' {
				break
			}
		}
	}

	buf.WriteByte('\n')

	return buf.String()
}
