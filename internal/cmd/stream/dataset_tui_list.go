package stream

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type dataset struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Desc      string    `json:"description"`
	CreatedBy string    `json:"who"`
	CreatedAt time.Time `json:"created"`
}

type DatasetsListModel struct {
	list list.Model
	Ctx  context.Context
	Opts *options
}

func (m DatasetsListModel) Init() tea.Cmd {
	return nil
}

func (m DatasetsListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if msg.String() == "q" {
			return m, tea.Quit
		}
		if msg.String() == "enter" {
			a := m.list.Items()[m.list.Index()]
			m.Opts.Dataset = a.(dataset).Name
			d := NewCharmDatasetsStream(m.Ctx, m.Opts)

			return d, nil
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m DatasetsListModel) View() string {
	return docStyle.Render(m.list.View())
}

func (d dataset) Title() string       { return fmt.Sprintf("%v (ID: %v)", d.Name, d.ID) }
func (d dataset) Description() string { return d.Desc }
func (d dataset) FilterValue() string { return d.Name }

func NewDatasetsList(ctx context.Context, opts *options) tea.Model {
	client, err := opts.Client(ctx)
	if err != nil {
		panic(err)
	}

	datasets, err := client.Datasets.List(ctx)
	if err != nil {
		panic(err)
	}
	// convert datasets to json then convert that json to list mode

	ds := []list.Item{}

	for _, d := range datasets {
		ds = append(ds, dataset{
			ID:        d.ID,
			Name:      d.Name,
			Desc:      d.Description,
			CreatedBy: d.CreatedBy,
			CreatedAt: d.CreatedAt,
		})
	}

	l := list.New(ds, list.NewDefaultDelegate(), 0, 0)

	return DatasetsListModel{l, ctx, opts}

}
