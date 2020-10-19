package printer

import (
	"math"
	"time"

	swagger "axicode.axiom.co/watchmakers/axiomdb/client/swagger/datasets"
	"github.com/TylerBrock/colorjson"
	"github.com/cli/cli/pkg/text"

	"github.com/axiomhq/cli/pkg/terminal"
)

// Table prints a table of entries.
func Table(io *terminal.IO, entries []swagger.Entry, withHeader bool) error {
	cs := io.ColorScheme()
	tp := terminal.NewTablePrinter(io)

	if io.IsStdoutTTY() && withHeader {
		tp.AddField("Time", cs.Bold)
		tp.AddField("Data", cs.Bold)
		tp.EndRow()
		tp.AddField("", nil)
		tp.EndRow()
	}

	for _, entry := range entries {
		tp.AddField(entry.Time.Format(time.RFC1123), cs.Gray)
		tp.AddField(jsonTransform(entry.Data))
		tp.EndRow()
	}
	return tp.Render()
}

// HINT(lukasmalkmus): Okay, this is hacky and kinda complicated, but here we
// go: table.AddField takes the text to render and a ColorFunc which colors the
// output. The text can't be colored beforehand for a reason: To calculate the
// width of a column, it needs the uncolored text because otherwise it would
// count the ASCII escape sequences as well, making the column way to large for
// the colored text, which is rendered with the same width as the uncolored one.
// So far so good. But internally, the text is also truncated before coloring
// it. This is no problem for normally colored text, but to color JSON
// correctly, it must be syntactically correct. Since simply truncating doesn't
// preserve the syntax, we have to do the stuff below:
// 1. Marshal to original text.
// 2. Marshal to colored text.
// 3. Return a ColorFunc that ignores the input given to it (which would be the
//    truncated text) and truncates the colored text, keeping in mind that we
//    need to truncate later because truncating counts the ANSII escape
//    sequences as well. Therefore we calculate the following:
//    i.) (length of untruncated, colored text) / (length of untruncated,
//        uncolored text): This gives us a float that describes the weight added
//        by the additional ANSII escape sequences.
//    ii.) Multiply that modifier with the length of the truncated, uncolored
//         text. This doesn't take into account that the untruncated, colored
//         text has additional spaces added, but it roughly works. We use the
// 		   resulting value to truncate the colored text.
func jsonTransform(data map[string]interface{}) (string, terminal.ColorFunc) {
	f := colorjson.NewFormatter()
	f.Indent = 0

	f.DisabledColor = true

	b, err := f.Marshal(data)
	if err != nil {
		return "", nil
	}

	f.DisabledColor = false

	cb, err := colorjson.Marshal(data)
	if err != nil {
		return "", nil
	}

	colorFunc := func(s string) string {
		mod := float64(len(cb)) / float64(len(b))
		colorWidth := int(math.Round(float64(len(s)) * mod))
		return text.Truncate(colorWidth, string(cb))
	}

	return string(b), colorFunc
}
