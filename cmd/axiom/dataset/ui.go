package dataset

import (
	"context"
	"fmt"

	axiomdb "axicode.axiom.co/watchmakers/axiomdb/client"
	swagger "axicode.axiom.co/watchmakers/axiomdb/client/swagger/datasets"
	pager "github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/indent"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/internal/config"
	"github.com/axiomhq/cli/pkg/terminal"
	"github.com/axiomhq/cli/pkg/utils"
	"github.com/axiomhq/cli/pkg/version"
)

const datasetsPerPage = 4

type state int

const (
	stateLoading state = iota
	stateNormal
	stateDeletingDataset
	stateQuitting
)

type datasetState int

const (
	datasetNormal datasetState = iota
	datasetSelected
	datasetDeleting
)

type datasetsLoadedMsg []swagger.Dataset

type deletedDatasetMsg int

type errMsg struct {
	err error
}

// Model is the Tea state model for the dataset user interface.
type Model struct {
	factory *cmdutil.Factory

	// Just shorthands so they don't need to be accessed through
	// model.factory each time.
	config *config.Config
	io     *terminal.IO
	cs     *terminal.ColorScheme

	err error

	standalone   bool
	state        state
	datasets     []swagger.Dataset
	datasetIndex int // Index of the selected dataset
	index        int // Index of selected dataset in relation to the current page

	Exit bool
	Quit bool

	spinner spinner.Model
	pager   pager.Model
}

// NewModel creates a new Tea state model for the dataset user interface.
func NewModel(f *cmdutil.Factory) *Model {
	m := &Model{
		factory: f,

		config: f.Config,
		io:     f.IO,
		cs:     f.IO.ColorScheme(),

		state:        stateLoading,
		datasets:     []swagger.Dataset{},
		datasetIndex: -1,

		spinner: spinner.NewModel(),
		pager:   pager.NewModel(),
	}

	m.spinner.Frames = spinner.Dot
	m.spinner.ForegroundColor = terminal.Magenta.Color(m.cs.IsDark())

	m.pager.PerPage = datasetsPerPage
	m.pager.Type = pager.Dots
	m.pager.InactiveDot = m.cs.Gray("•")

	return m
}

// Init is the Tea initialization function.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.listDatasets(),
		spinner.Tick(m.spinner),
	)
}

// Update is the tea update function which handles incoming messages.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			m.state = stateQuitting
			return m, tea.Quit
		}

		switch msg.String() {
		case "q", "esc":
			if m.standalone {
				m.state = stateQuitting
				return m, tea.Quit
			}
			m.Exit = true
			return m, nil

		// Select individual items
		case "up", "k":
			// Move up
			m.index--
			if m.index < 0 && m.pager.Page > 0 {
				m.index = m.pager.PerPage - 1
				m.pager.PrevPage()
			}
			m.index = max(0, m.index)
		case "down", "j":
			// Move down
			itemsOnPage := m.pager.ItemsOnPage(len(m.datasets))
			m.index++
			if m.index > itemsOnPage-1 && m.pager.Page < m.pager.TotalPages-1 {
				m.index = 0
				m.pager.NextPage()
			}
			m.index = min(itemsOnPage-1, m.index)

		// Delete
		case "x":
			m.state = stateDeletingDataset
			m.updatePaging(msg)
			return m, nil

			// Confirm Delete
		case "y":
			switch m.state {
			case stateDeletingDataset:
				if m.getSelectedIndex() == m.index {
					// The user is going to delete
					m.state = stateDeletingDataset
					return m, nil
				}
				m.state = stateNormal
				return m, m.deleteDataset()
			}
		}

	case errMsg:
		m.err = msg.err
		return m, nil

	case datasetsLoadedMsg:
		m.state = stateNormal
		m.index = 0
		m.datasets = msg

	case deletedDatasetMsg:
		if m.state == stateQuitting {
			return m, tea.Quit
		}
		i := m.getSelectedIndex()

		// Remove dataset from slice.
		m.datasets = append(m.datasets[:i], m.datasets[i+1:]...)

		// Update pagination.
		m.pager.SetTotalPages(len(m.datasets))
		m.pager.Page = min(m.pager.Page, m.pager.TotalPages-1)

		// Update cursor.
		m.index = min(m.index, m.pager.ItemsOnPage(len(m.datasets)-1))

		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		if m.state < stateNormal {
			m.spinner, cmd = spinner.Update(msg, m.spinner)
		}
		return m, cmd
	}

	m.updatePaging(msg)

	// If an item is being confirmed for delete, any key (other than the key
	// used for confirmation above) cancels the deletion
	k, ok := msg.(tea.KeyMsg)
	if ok && k.String() != "x" {
		m.state = stateNormal
	}

	return m, nil
}

// View renders the current UI into a string.
func (m *Model) View() string {
	if m.err != nil {
		return m.err.Error()
	}

	var s string

	switch m.state {
	case stateLoading:
		s = spinner.View(m.spinner) + " Loading...\n\n"
	case stateQuitting:
		s = "Thanks for using Axiom!\n"
	default:
		s = fmt.Sprintf("Showing %s for backend %s:\n\n",
			utils.Pluralize(m.cs, "dataset", len(m.datasets)),
			m.cs.Bold(m.config.ActiveBackend),
		)

		// Datasets
		s += m.datasetsView()
		if m.pager.TotalPages > 1 {
			s += pager.View(m.pager)
		}

		// Footer
		switch m.state {
		case stateDeletingDataset:
			s += m.promptDeleteView()
		default:
			s += m.helpView()
		}
	}

	if m.standalone {
		return indent.String(fmt.Sprintf("\n%s\n", s), 2)
	}
	return s
}

// updatePaging runs an update against the underlying pagination model as well
// as performing some related tasks on this model.
func (m *Model) updatePaging(msg tea.Msg) {
	// Handle paging
	m.pager.SetTotalPages(len(m.datasets))
	m.pager, _ = pager.Update(msg, m.pager)

	// If selected item is out of bounds, put it in bounds
	numItems := m.pager.ItemsOnPage(len(m.datasets))
	m.index = min(m.index, numItems-1)
}

// getSelectedIndex returns the index of the cursor in relation to the total
// number of items.
func (m *Model) getSelectedIndex() int {
	return m.index + m.pager.Page*m.pager.PerPage
}

func (m *Model) datasetsView() string {
	var (
		s          string
		state      datasetState
		start, end = m.pager.GetSliceBounds(len(m.datasets))
		datasets   = m.datasets[start:end]
	)

	// Render dataset info
	for i, dataset := range datasets {
		if m.state == stateDeletingDataset && m.index == i {
			state = datasetDeleting
		} else if m.index == i {
			state = datasetSelected
		} else {
			state = datasetNormal
		}
		s += newStyledDataset(m.cs, dataset, i+start == m.datasetIndex).render(m.cs, state)
	}

	// If there aren't enough datasets to fill the view, fill the missing parts
	// with whitespace
	if len(datasets) < m.pager.PerPage {
		for i := len(datasets); i < datasetsPerPage; i++ {
			s += "\n\n\n"
		}
	}

	return s
}

func (m *Model) listDatasets() tea.Cmd {
	return func() tea.Msg {
		client, err := m.factory.Client()
		if err != nil {
			return errMsg{err}
		}
		datasets, err := client.Datasets.List(context.Background(), axiomdb.ListOptions{})
		if err != nil {
			return errMsg{err}
		}
		return datasetsLoadedMsg(datasets)
	}
}

func (m *Model) deleteDataset() tea.Cmd {
	return func() tea.Msg {
		client, err := m.factory.Client()
		if err != nil {
			return errMsg{err}
		} else if err = client.Datasets.Delete(context.Background(), m.datasets[m.datasetIndex].Name); err != nil {
			return errMsg{err}
		}
		return deletedDatasetMsg(m.index)
	}
}

func (m *Model) promptDeleteView() string {
	return m.cs.Red("\n\nDelete this dataset? ") + m.cs.Gray("(y/N)")
}

func (m *Model) helpView() string {
	return "\n\n" + m.cs.Gray("Axiom CLI "+version.Release()+" • j/k, ↑/↓: choose • h/l, ←/→: page • x: delete • esc: exit")
}

// Utils

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
