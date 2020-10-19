package dataset

import (
	"fmt"
	"strconv"
	"time"

	swagger "axicode.axiom.co/watchmakers/axiomdb/client/swagger/datasets"
	"github.com/charmbracelet/charm/ui/common"
	te "github.com/muesli/termenv"

	"github.com/axiomhq/cli/pkg/terminal"
)

type styledDataset struct {
	swagger.Dataset

	id        string
	name      string
	createdAt string
	note      string
	gutter    string
}

func newStyledDataset(cs *terminal.ColorScheme, ds swagger.Dataset, active bool) styledDataset {
	var note string
	if active {
		bullet := te.String("• ").Foreground(common.NewColorPair("#2B4A3F", "#ABE5D1").Color()).String()
		note = bullet + te.String("Current Key").Foreground(common.NewColorPair("#04B575", "#04B575").Color()).String()
	}

	return styledDataset{
		Dataset: ds,

		id:        cs.White("ID: ") + strconv.Itoa(int(ds.ID)),
		name:      cs.White("Name: ") + ds.Name,
		createdAt: cs.White("Create: ") + ds.CreatedAt.Format(time.RFC1123),
		note:      note,
		gutter:    " ",
	}
}

func (ds *styledDataset) selected(cs *terminal.ColorScheme) {
	ds.gutter = cs.Magenta("│")
	ds.id = cs.Magenta("ID: ") + cs.Cyan(strconv.Itoa(int(ds.ID)))
	ds.name = cs.Magenta("Name: ") + cs.Cyan(ds.Name)
	ds.createdAt = cs.Magenta("Create: ") + cs.Cyan(ds.CreatedAt.Format(time.RFC1123))
}

func (ds *styledDataset) deleting(cs *terminal.ColorScheme) {
	ds.gutter = cs.Red("│")
	ds.id = cs.Red("ID: " + strconv.Itoa(int(ds.ID)))
	ds.name = cs.Red("Name: " + ds.Name)
	ds.createdAt = cs.Red("Create: " + ds.CreatedAt.Format(time.RFC1123))
}

func (ds styledDataset) render(cs *terminal.ColorScheme, state datasetState) string {
	switch state {
	case datasetSelected:
		ds.selected(cs)
	case datasetDeleting:
		ds.deleting(cs)
	}
	return fmt.Sprintf(
		"%s %s\n%s %s\n%s %s %s\n\n",
		ds.gutter, ds.id,
		ds.gutter, ds.name,
		ds.gutter, ds.createdAt, ds.note,
	)
}
