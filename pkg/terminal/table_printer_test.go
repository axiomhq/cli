// https://github.com/cli/cli/blob/trunk/utils/table_printer_test.go
package terminal

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ttyTablePrinter_truncate(t *testing.T) {
	buf := bytes.Buffer{}
	tp := &ttyTablePrinter{
		out:      &buf,
		maxWidth: 5,
	}

	tp.AddField("1", nil)
	tp.AddField("hello", nil)
	tp.EndRow()
	tp.AddField("2", nil)
	tp.AddField("world", nil)
	tp.EndRow()

	err := tp.Render()
	assert.NoError(t, err)

	assert.Equal(t, "1  he\n2  wo\n", buf.String())
}
