package stream

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	axiomdb "axicode.axiom.co/watchmakers/axiomdb/client"
	swagger "axicode.axiom.co/watchmakers/axiomdb/client/swagger/datasets"
	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/text"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
)

const streamingDuration = time.Second * 3

type options struct {
	*cmdutil.Factory

	// Dataset to stream from. If not supplied as an argument, which is
	// optional, the user will be asked for it.
	Dataset string
}

// NewStreamCmd creates and returns the stream command.
func NewStreamCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &options{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "stream [(-d|--dataset) <dataset-name>]",
		Short: "Livestream data",
		Long:  `Livestream data from an Axiom dataset.`,

		DisableFlagsInUseLine: true,

		Example: heredoc.Doc(`
			# Interactively stream a dataset:
			$ axiom stream
						
			# Stream the "logs" dataset:
			$ axiom stream -d logs
		`),

		Annotations: map[string]string{
			"IsCore": "true",
		},

		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := cmdutil.Needs(
				cmdutil.NeedsActiveBackend(f),
				cmdutil.NeedsDatasets(f),
			)(cmd, args); err != nil {
				return err
			}

			if opts.IO.IsStdinTTY() {
				return complete(cmd.Context(), opts)
			}
			return nil
		},

		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Dataset, "dataset", "d", "", "Dataset to stream")

	_ = cmd.RegisterFlagCompletionFunc("dataset", cmdutil.DatasetCompletionFunc(f))

	if !opts.IO.IsStdinTTY() {
		_ = cmd.MarkFlagRequired("dataset")
	}

	return cmd
}

func complete(ctx context.Context, opts *options) error {
	if opts.Dataset != "" {
		return nil
	}

	client, err := opts.Client()
	if err != nil {
		return err
	}

	stop := opts.IO.StartProgressIndicator()
	defer stop()

	datasets, err := client.Datasets.List(ctx, axiomdb.ListOptions{})
	if err != nil {
		return err
	}

	stop()

	datasetNames := make([]string, len(datasets))
	for i, dataset := range datasets {
		datasetNames[i] = dataset.Name
	}

	return survey.AskOne(&survey.Select{
		Message: "Which dataset to stream from?",
		Options: datasetNames,
	}, &opts.Dataset, opts.IO.SurveyIO())
}

func run(ctx context.Context, opts *options) error {
	if !opts.IO.IsStdoutTTY() || !opts.IO.IsStderrTTY() || !opts.IO.ColorEnabled() {
		return errors.New("can only run streaming mode in TTY mode with color support")
	}

	client, err := opts.Client()
	if err != nil {
		return err
	}

	term, err := termbox.New()
	if err != nil {
		return err
	}
	defer term.Close()

	stream, err := text.New(
		text.RollContent(),
		text.ScrollKeys(
			keyboard.KeyArrowUp,
			keyboard.KeyArrowDown,
			keyboard.KeyPgUp,
			keyboard.KeyPgDn,
		),
	)
	if err != nil {
		return err
	}
	_ = printFormatted(stream, cell.ColorMagenta, "Streaming events from %s\n\n", opts.Dataset)

	ui, err := container.New(
		term,
		container.Border(linestyle.Round),
		container.BorderTitle("AXIOM CLI UI -- PRESS Q TO QUIT"),
		container.BorderTitleAlignCenter(),
		container.PlaceWidget(stream),
	)
	if err != nil {
		return err
	}

	// Extra context so we can cancel on keyboard button press.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		t := time.NewTicker(streamingDuration)
		defer t.Stop()

		lastRequest := time.Now().Add(-time.Nanosecond)
		for {
			queryCtx, queryCancel := context.WithTimeout(ctx, streamingDuration)

			res, err := client.Datasets.Query(queryCtx, opts.Dataset, swagger.QueryRequest{
				StartTime: lastRequest,
				EndTime:   time.Now(),
			}, axiomdb.QueryOptions{
				StreamingDuration: streamingDuration,
			})
			if err != nil && !errors.Is(err, context.DeadlineExceeded) {
				if !errors.Is(err, context.Canceled) {
					_ = printFormatted(stream, cell.ColorRed, "Streaming error: %s\n", err)
				}
			} else if res != nil {
				if len(res.Matches) > 0 {
					lastRequest = res.Matches[len(res.Matches)-1].Time.Add(time.Nanosecond)
				}

				for _, entry := range res.Matches {
					_ = printEntry(stream, entry)
				}
			}

			queryCancel()

			select {
			case <-ctx.Done():
				return
			case <-t.C:
			}
		}
	}()

	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == 'q' || k.Key == 'Q' {
			cancel()
		}
	}

	return termdash.Run(ctx, term, ui, termdash.KeyboardSubscriber(quitter))
}

func printFormatted(txt *text.Text, cl cell.Color, format string, a ...interface{}) error {
	s := fmt.Sprintf(format, a...)
	return txt.Write(s, text.WriteCellOpts(cell.FgColor(cl)))
}

const (
	keyColor    = cell.ColorWhite
	stringColor = cell.ColorGreen
	boolColor   = cell.ColorYellow
	numberColor = cell.ColorCyan
	nullColor   = cell.ColorMagenta
)

// Derived from https://github.com/cli/cli/blob/trunk/pkg/jsoncolor/jsoncolor.go.
func printEntry(txt *text.Text, entry swagger.Entry) error {
	// HINT(lukasmalkmus): Yes, this is kinda hacky because we marshal to JSON
	// what we previously unmarshalled fronm JSON. But int he future, we should
	// just read the raw response and have the entry contain the raw JSON event
	// data.

	b, err := json.Marshal(entry.Data)
	if err != nil {
		return err
	}

	_ = txt.Write(entry.Time.Format(time.RFC1123))
	_ = txt.Write("  ")

	r := bytes.NewBuffer(b)
	dec := json.NewDecoder(r)
	dec.UseNumber()

	var (
		idx   int
		stack []json.Delim
	)

	for {
		t, err := dec.Token()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		switch tt := t.(type) {
		case json.Delim:
			switch tt {
			case '{', '[':
				stack = append(stack, tt)
				idx = 0
				_ = txt.Write(tt.String(), text.WriteCellOpts(cell.FgColor(keyColor)))
				continue
			case '}', ']':
				stack = stack[:len(stack)-1]
				idx = 0
				_ = txt.Write(tt.String(), text.WriteCellOpts(cell.FgColor(keyColor)))
			}
		default:
			b, err := json.Marshal(tt)
			if err != nil {
				return err
			}

			isKey := len(stack) > 0 && stack[len(stack)-1] == '{' && idx%2 == 0
			idx++

			var color cell.Color
			if isKey {
				color = keyColor
			} else if tt == nil {
				color = nullColor
			} else {
				switch t.(type) {
				case string:
					color = stringColor
				case bool:
					color = boolColor
				case json.Number:
					color = numberColor
				}
			}

			_ = txt.Write(string(b), text.WriteCellOpts(cell.FgColor(color)))
			if isKey {
				_ = txt.Write(": ", text.WriteCellOpts(cell.FgColor(keyColor)))
				continue
			}
		}

		if dec.More() {
			_ = txt.Write(", ", text.WriteCellOpts(cell.FgColor(keyColor)))
		}
	}

	_ = txt.Write("\n", text.WriteCellOpts(cell.FgColor(keyColor)))

	return nil
}
