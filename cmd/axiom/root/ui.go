package root

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/indent"

	"github.com/axiomhq/cli/cmd/axiom/dataset"
	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/internal/config"
	"github.com/axiomhq/cli/pkg/terminal"
	"github.com/axiomhq/cli/pkg/version"
)

type state int

const (
	stateInit state = iota
	stateReady
	// stateBrowsingAuth
	// stateBrowsingConfig
	stateBrowsingDataset
	// stateBrowsingIngest
	// stateBrowsingIntegrate
	// stateBrowsingStream
	stateQuitting
)

type menuChoice int

const (
	choiceAuth menuChoice = iota
	choiceConfig
	choiceDataset
	choiceIngest
	choiceIntegrate
	choiceStream
	choiceExit
	choiceUnset
)

// TODO: Build helpers that generate those menu items from cobras command tree.

var menuChoices = map[menuChoice]string{
	choiceAuth:      "Manage authentication state",
	choiceConfig:    "Manage configuration",
	choiceDataset:   "Manage datasets",
	choiceIngest:    "Ingest data",
	choiceIntegrate: "Integrate Axiom into a project",
	choiceStream:    "Livestream data",
	choiceExit:      "Exit",
}

type model struct {
	factory *cmdutil.Factory

	// Just shorthands so they don't need to be accessed through
	// model.factory each time.
	config *config.Config
	io     *terminal.IO
	cs     *terminal.ColorScheme

	state      state
	menuIndex  int
	menuChoice menuChoice

	spinner spinner.Model
	dataset *dataset.Model
}

func newModel(f *cmdutil.Factory) *model {
	m := &model{
		factory: f,

		config: f.Config,
		io:     f.IO,
		cs:     f.IO.ColorScheme(),

		menuChoice: choiceUnset,

		spinner: spinner.NewModel(),
		dataset: dataset.NewModel(f),
	}

	m.spinner.Frames = spinner.Dot
	m.spinner.ForegroundColor = terminal.Magenta.Color(m.cs.IsDark())

	return m
}

// Init implements bubbletea.Model.
func (m *model) Init() tea.Cmd {
	m.state = stateReady
	return spinner.Tick(m.spinner)
}

// Update implements bubbletea.Model.
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			m.state = stateQuitting
			return m, tea.Quit
		}

		// Process keys for the menu
		if m.state == stateReady {
			switch msg.String() {
			// Quit
			case "q", "esc":
				m.state = stateQuitting
				return m, tea.Quit

			// Prev menu item
			case "up", "k":
				m.menuIndex--
				if m.menuIndex < 0 {
					m.menuIndex = len(menuChoices) - 1
				}

			// Select menu item
			case "enter":
				m.menuChoice = menuChoice(m.menuIndex)

			// Next menu item
			case "down", "j":
				m.menuIndex++
				if m.menuIndex >= len(menuChoices) {
					m.menuIndex = 0
				}
			}
		}

	case spinner.TickMsg:
		switch m.state {
		case stateInit:
			var cmd tea.Cmd
			m.spinner, cmd = spinner.Update(msg, m.spinner)
			return m, cmd
		}
	}

	return m, m.updateChilden(msg)
}

func (m *model) updateChilden(_ tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	switch m.state {
	case stateBrowsingDataset:
		// newModel, newCmd := m.dataset.Update(msg)
		// datasetModel, ok := newModel.(dataset.Model)
		// if !ok {
		// 	panic("could not perform assertion on dataset model")
		// }
		// m.dataset = datasetModel
		// cmd = newCmd

		// if m.dataset.Exit {
		// 	m.dataset = dataset.NewModel(m.factory)
		// 	m.state = stateReady
		// } else if m.dataset.Quit {
		// 	m.state = stateQuitting
		// 	return tea.Quit
		// }
	}

	// Handle the menu
	switch m.menuChoice {
	case choiceDataset:
		// m.state = stateBrowsingDataset
		// m.menuChoice = choiceUnset
		// cmd = dataset.LoadDatasets(m.dataset)

	case choiceExit:
		m.state = stateQuitting
		cmd = tea.Quit
	}

	return cmd
}

// Update implements bubbletea.Model.
func (m *model) View() string {
	s := m.logoView()

	switch m.state {
	case stateInit:
		s += spinner.View(m.spinner) + " Initializing..."
	case stateReady:
		s += m.infoView()
		s += m.menuView()
		s += m.footerView()
	case stateBrowsingDataset:
		// s += m.datasets.View()
	case stateQuitting:
		s += m.quitView()
	}

	return indent.String(s, 2) + "\n"
}

func (m *model) logoView() string {
	return "\n" + m.cs.Title(" Axiom ") + "\n\n"
}

func (m *model) infoView() string {
	var s string

	backend, ok := m.config.Backends[m.config.ActiveBackend]
	if !ok {
		s += "│ Active Backend: " + m.cs.Gray("(none set)")
	} else {
		s += m.cs.Green("│") + " Active Backend: " + m.cs.Magenta(m.config.ActiveBackend) + "\n"
		s += m.cs.Green("│") + " Username: " + m.cs.Magenta(backend.Username) + "\n"
		s += m.cs.Green("│") + " URL: " + m.cs.Magenta(backend.URL)
	}

	return s + "\n\n"
}

func (m *model) menuView() string {
	var s string
	for i := 0; i < len(menuChoices); i++ {
		e := "  "
		if i == m.menuIndex {
			e = m.cs.Magenta("> " + menuChoices[menuChoice(i)])
		} else {
			e += menuChoices[menuChoice(i)]
		}
		if i < len(menuChoices)-1 {
			e += "\n"
		}
		s += e
	}

	return s
}

func (m *model) quitView() string {
	return "Thanks for using Axiom!\n"
}

func (m *model) footerView() string {
	return "\n\n" + m.cs.Gray("Axiom CLI "+version.Release()+" • j/k, ↑/↓: choose • enter: select")
}
