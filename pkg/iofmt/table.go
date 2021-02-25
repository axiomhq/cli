package iofmt

import (
	"errors"
	"io"

	"github.com/axiomhq/cli/pkg/terminal"
)

// TableRowBuilder provides methods for building table rows. It is a
// sub-interface of terminal.TablePrinter.
type TableRowBuilder interface {
	// AddField adds a field with the given string to the row. If given, the
	// ColorFunc will be used to colorize the resulting field text.
	AddField(string, terminal.ColorFunc)
}

// HeaderBuilderFunc allows building of a tables header by providing a
// TableRowBuilder and the raw io.Writer for using generic formatting
// functions like fmt.Fprintf().
type HeaderBuilderFunc func(io.Writer, TableRowBuilder)

// RowBuilderFunc allows building of a tables content row by providing a
// TableRowBuilder and the id of the record to format.
type RowBuilderFunc func(TableRowBuilder, int)

// FormatToTable formats in Table format using the given BuilderFunc's. Header
// and footer builder functions are optional. The length of the slice must be
// given.
func FormatToTable(io *terminal.IO, l int, header, footer HeaderBuilderFunc, contentRow RowBuilderFunc) error {
	tp := terminal.NewTablePrinter(io)

	if contentRow == nil {
		return errors.New("missing table row content")
	}

	if header != nil {
		header(io.Out(), tp)

		// Empty row between header and content.
		tp.EndRow()
		tp.AddField("", nil)
		tp.EndRow()
	}

	if l > 1 {
		for k := 0; k < l; k++ {
			contentRow(tp, k)
			tp.EndRow()
		}
	} else {
		contentRow(tp, -1)
		tp.EndRow()
	}

	if footer != nil {
		// Empty row between content and footer.
		tp.AddField("", nil)
		tp.EndRow()

		footer(io.Out(), tp)
		tp.EndRow()
	}

	return tp.Render()
}
