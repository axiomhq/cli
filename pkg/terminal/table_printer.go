// https://github.com/cli/cli/blob/trunk/utils/table_printer.go
package terminal

import (
	"fmt"
	"io"
	"strings"

	"github.com/cli/cli/pkg/text"
)

// TablePrinter prints table formatted output.
type TablePrinter interface {
	// AddField adds a field with the given string to the row. If given, the
	// ColorFunc will be used to colorize the resulting field text.
	AddField(string, ColorFunc)
	// EndRow ends a row and starts with a new one on the next AddField call.
	EndRow()
	// Render the table to the underlying output. It also clears internal state
	// and makes the TablePrinter ready to use again.
	Render() error
}

// NewTablePrinter creates a new TablePrinter writing its output to the
// underlying IO. It auto detects if a TTY is attached and uses a different
// output format that is easy to parse when piped.
func NewTablePrinter(io *IO) TablePrinter {
	if io.isStdoutTTY {
		return &ttyTablePrinter{
			out:      io.out,
			maxWidth: io.TerminalWidth(),
		}
	}
	return &tsvTablePrinter{
		out: io.out,
	}
}

type tableField struct {
	Text         string
	TruncateFunc func(int, string) string
	ColorFunc    func(string) string
}

type ttyTablePrinter struct {
	out      io.Writer
	maxWidth int
	rows     [][]tableField
}

// AddField adds a field with the given string to the row. If given, the
// ColorFunc will be used to colorize the resulting field text.
func (t *ttyTablePrinter) AddField(s string, colorFunc ColorFunc) {
	if t.rows == nil {
		t.rows = make([][]tableField, 1)
	}
	rowI := len(t.rows) - 1
	field := tableField{
		Text:         s,
		TruncateFunc: text.Truncate,
		ColorFunc:    colorFunc,
	}
	t.rows[rowI] = append(t.rows[rowI], field)
}

// EndRow ends a row and starts with a new one on the next AddField call.
func (t *ttyTablePrinter) EndRow() {
	t.rows = append(t.rows, []tableField{})
}

// Render the table to the underlying output.
func (t *ttyTablePrinter) Render() error {
	if len(t.rows) == 0 {
		return nil
	}

	numCols := len(t.rows[0])
	colWidths := make([]int, numCols)
	// measure maximum content width per column
	for _, row := range t.rows {
		for col, field := range row {
			textLen := text.DisplayWidth(field.Text)
			if textLen > colWidths[col] {
				colWidths[col] = textLen
			}
		}
	}

	delim := "  "
	availWidth := t.maxWidth - colWidths[0] - ((numCols - 1) * len(delim))
	// add extra space from columns that are already narrower than threshold
	for col := 1; col < numCols; col++ {
		availColWidth := availWidth / (numCols - 1)
		if extra := availColWidth - colWidths[col]; extra > 0 {
			availWidth += extra
		}
	}
	// cap all but first column to fit available terminal width
	// TODO: support weighted instead of even redistribution
	for col := 1; col < numCols; col++ {
		availColWidth := availWidth / (numCols - 1)
		if colWidths[col] > availColWidth {
			colWidths[col] = availColWidth
		}
	}

	for _, row := range t.rows {
		for col, field := range row {
			if col > 0 {
				if _, err := fmt.Fprint(t.out, delim); err != nil {
					return err
				}
			}
			truncVal := field.TruncateFunc(colWidths[col], field.Text)
			if col < numCols-1 {
				// pad value with spaces on the right
				if padWidth := colWidths[col] - text.DisplayWidth(field.Text); padWidth > 0 {
					truncVal += strings.Repeat(" ", padWidth)
				}
			}
			if field.ColorFunc != nil {
				truncVal = field.ColorFunc(truncVal)
			}
			if _, err := fmt.Fprint(t.out, truncVal); err != nil {
				return err
			}
		}
		if len(row) > 0 {
			if _, err := fmt.Fprint(t.out, "\n"); err != nil {
				return err
			}
		}
	}
	t.rows = nil
	return nil
}

type tsvTablePrinter struct {
	out        io.Writer
	currentCol int
}

// AddField adds a field with the given string to the row. The ColorFunc is
// ignored.
func (t *tsvTablePrinter) AddField(s string, _ ColorFunc) {
	if t.currentCol > 0 {
		fmt.Fprint(t.out, "\t")
	}
	fmt.Fprint(t.out, s)
	t.currentCol++
}

// EndRow ends a row and starts with a new one on the next AddField call.
func (t *tsvTablePrinter) EndRow() {
	fmt.Fprint(t.out, "\n")
	t.currentCol = 0
}

// Render the table to the underlying output.
func (t *tsvTablePrinter) Render() error {
	t.currentCol = 0
	return nil
}
