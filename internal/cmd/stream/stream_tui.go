package stream

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/axiom/query"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type DatasetsStreamModel struct {
	ctx     context.Context
	errChan chan error
	opts    *options
	FullStr string
	client  *axiom.Client

	dataStream chan []query.Entry
}

type StateMsg struct {
	running bool
}

type ConnectedMsg struct {
}

func StartCharmDatasetsStream(ctx context.Context, opts *options) error {
	if opts.Dataset == "" {
		d := NewDatasetsList(ctx, opts)
		p := tea.NewProgram(d)
		if err := p.Start(); err != nil {
			return err
		}
	} else {

		m := NewCharmDatasetsStream(ctx, opts)
		p := tea.NewProgram(m)
		if err := p.Start(); err != nil {
			return err
		}
	}

	return nil
}

func NewCharmDatasetsStream(ctx context.Context, opts *options) DatasetsStreamModel {
	fmt.Println("Connecting to Axiom...")
	client, err := opts.Client(ctx)
	if err != nil {
		panic(err)
	}
	return DatasetsStreamModel{
		ctx:        ctx,
		errChan:    make(chan error),
		opts:       opts,
		dataStream: make(chan []query.Entry),
		client:     client,
	}
}

func (m *DatasetsStreamModel) tick(d time.Duration) tea.Cmd {
	var err error

	lastRequest := time.Now().Add(-time.Nanosecond)

	queryCtx, queryCancel := context.WithTimeout(m.ctx, streamingDuration)

	res, err := m.client.Datasets.Query(queryCtx, m.opts.Dataset, query.Query{
		StartTime: lastRequest,
		EndTime:   time.Now(),
	}, query.Options{
		StreamingDuration: streamingDuration,
	})
	if err != nil && !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		queryCancel()
		m.errChan <- err
	}

	queryCancel()

	if res != nil && len(res.Matches) > 0 {
		lastRequest = res.Matches[len(res.Matches)-1].Time.Add(time.Nanosecond)
		m.dataStream <- res.Matches
	}

	return tea.Tick(d, func(_ time.Time) tea.Msg {
		return StateMsg{running: true}
	})
}

func (m DatasetsStreamModel) Init() tea.Cmd {
	fmt.Println("Init")
	return tea.Batch(func() tea.Msg {
		m.FullStr = docStyle.Render("Streaming...")
		return StateMsg{running: true}
	}, m.tick(time.Second))
}

func (m DatasetsStreamModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}
		return m, nil
	case StateMsg:
		return m, m.tick(time.Second)
	case ConnectedMsg:
		m.FullStr = docStyle.Render("Connected, streaming...")
		return m, m.tick(time.Millisecond)
	}

	return m, nil
}

func (m DatasetsStreamModel) View() string {
	select {
	case err := <-m.errChan:
		fmt.Printf("Error: %s", err)
		tea.Quit()
		return ""
	case entry := <-m.dataStream:
		for _, e := range entry {
			m.FullStr += docStyle.Render(fmt.Sprintf("%s: %s", e.Time, e.Data)) + "\n"
		}
		return m.FullStr
	default:
		return m.FullStr
	}
}
