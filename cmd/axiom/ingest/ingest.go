package ingest

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"
	"unicode"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/axiomhq/axiom-go/axiom"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/pkg/utils"
)

type options struct {
	*cmdutil.Factory

	// Dataset to ingest into. If not supplied as an argument, which is
	// optional, the user will be asked for it.
	Dataset string
	// Filenames of the files to ingest. If not set, will read from Stdin.
	Filenames []string
	// TimestampField to take the ingestion time from.
	TimestampField string
	// TimestampFormat the timestamp is formatted in.
	TimestampFormat string
	// FlushEvery flushes the ingestion buffer after the specified duration. It
	// is only valid when ingesting a stream of newline delimited JSON objects
	// of unknown length.
	FlushEvery time.Duration
	// Compression enables gzip compression.
	Compression bool
}

// NewIngestCmd creates and returns the ingest command.
func NewIngestCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &options{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "ingest <dataset-name> (-|(-f|--file) <filename>...) [--timestamp-field <timestamp-field>] [--timestamp-format <timestamp-format>] [--flush-every <duration>] [(-c|--compression=)TRUE|FALSE]",
		Short: "Ingest data",
		Long:  `Ingest data into an Axiom dataset.`,

		DisableFlagsInUseLine: true,

		Args:              cmdutil.PopulateFromArgs(f, &opts.Dataset),
		ValidArgsFunction: cmdutil.DatasetCompletionFunc(f),

		Example: heredoc.Doc(`
			# Ingest the contents of a JSON logfile into a dataset named
			# "nginx-logs":
			$ axiom ingest nginx-logs -f nginx-logs.json

			# Ingest the contents of all files inside /var/logs/nginx with
			# extension ".log" into a dataset named "nginx-logs":
			$ axiom ingest nginx-logs -f /var/logs/nginx/*.log

			# Pipe the contents of a log generator into a dataset named
			# "gen-logs". If the length of the data stream is unknown, the
			# "--flush-every" flag must be passed to ship the data to the server
			# after the specified duration instead of waiting for EOF. This is
			# only valid for newline delimited JSON.
			$ ./loggen -ndjson | axiom ingest gen-logs
		`),

		Annotations: map[string]string{
			"IsCore": "true",
		},

		PreRunE: cmdutil.ChainRunFuncs(
			cmdutil.NeedsActiveDeployment(f),
			cmdutil.NeedsDatasets(f),
		),

		RunE: func(cmd *cobra.Command, args []string) error {
			// When no files are specified, stdin is the file to use.
			if len(opts.Filenames) == 0 {
				opts.Filenames = []string{"-"}
			}

			if err := complete(cmd.Context(), opts); err != nil {
				return err
			}
			return run(cmd.Context(), opts, cmd.Flag("flush-every").Changed)
		},
	}

	cmd.Flags().StringSliceVarP(&opts.Filenames, "file", "f", nil, "File to ingest")
	cmd.Flags().StringVar(&opts.TimestampField, "timestamp-field", "", "Field to take the ingestion time from")
	cmd.Flags().StringVar(&opts.TimestampFormat, "timestamp-format", "", "Format the timestamp is formatted in")
	cmd.Flags().DurationVar(&opts.FlushEvery, "flush-every", time.Second, "Buffer flush interval for newline delimited JSON streams of unknown length")
	cmd.Flags().BoolVarP(&opts.Compression, "compression", "c", true, "Enable gzip compression")

	_ = cmd.RegisterFlagCompletionFunc("timestamp-field", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("timestamp-format", cmdutil.NoCompletion)

	if opts.IO.IsStdinTTY() {
		_ = cmd.MarkFlagRequired("file")
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

	stop := opts.IO.StartActivityIndicator()
	datasets, err := client.Datasets.List(ctx)
	if err != nil {
		stop()
		return err
	}
	stop()

	datasetNames := make([]string, len(datasets))
	for i, dataset := range datasets {
		datasetNames[i] = dataset.Name
	}

	return survey.AskOne(&survey.Select{
		Message: "Which dataset to ingest into?",
		Options: datasetNames,
	}, &opts.Dataset, opts.IO.SurveyIO())
}

func run(ctx context.Context, opts *options, flushEverySet bool) error {
	client, err := opts.Client()
	if err != nil {
		return err
	}

	stop := opts.IO.StartActivityIndicator()

	var (
		res     = new(axiom.IngestStatus)
		lastErr error
	)
	for _, filename := range opts.Filenames {
		var rc io.ReadCloser
		if filename == "-" {
			rc = opts.IO.In()
			filename = "stdin" // Enhance printed output
		} else {
			if rc, err = os.Open(filename); err != nil {
				lastErr = err
				break
			}
		}

		var (
			r   io.Reader
			typ axiom.ContentType
		)
		if r, typ, err = detectContentType(rc); err != nil {
			_ = rc.Close()
			lastErr = fmt.Errorf("could not detect %q content type: %w", filename, err)
			break
		}

		if flushEverySet && typ != axiom.NDJSON {
			return errors.New("--flush-every not valid when content type is not newline delimited JSON")
		}

		var ingestRes *axiom.IngestStatus
		if filename == "stdin" && typ == axiom.NDJSON {
			ingestRes, err = ingestEvery(ctx, client, r, opts)
		} else {
			ingestRes, err = ingest(ctx, client, r, typ, opts)
		}
		mergeIngestStatuses(res, ingestRes)

		if err != nil && !errors.Is(err, context.Canceled) {
			_ = rc.Close()
			lastErr = fmt.Errorf("could not ingest %q into dataset %q: %w", filename, opts.Dataset, err)
			break
		} else if err = rc.Close(); err != nil {
			lastErr = fmt.Errorf("failed to close %q: %w", filename, err)
			break
		} else if errors.Is(err, context.Canceled) {
			break
		}
	}

	stop()

	if opts.IO.IsStderrTTY() {
		cs := opts.IO.ColorScheme()

		if res.Ingested > 0 {
			fmt.Fprintf(opts.IO.ErrOut(), "%s Ingested %s (%s)\n",
				cs.SuccessIcon(),
				utils.Pluralize(cs, "event", int(res.Ingested)),
				humanize.Bytes(res.ProcessedBytes),
			)
		}

		if res.Failed > 0 {
			fmt.Fprintf(opts.IO.ErrOut(), "%s Failed to ingest %s:\n\n",
				cs.ErrorIcon(),
				utils.Pluralize(cs, "event", int(res.Failed)),
			)
			for _, fail := range res.Failures {
				fmt.Fprintf(opts.IO.ErrOut(), "%s: %s\n",
					cs.Gray(fail.Timestamp.Format(time.RFC1123)), err,
				)
			}
		}
	}

	return lastErr
}

func ingestEvery(ctx context.Context, client *axiom.Client, r io.Reader, opts *options) (*axiom.IngestStatus, error) {
	t := time.NewTicker(opts.FlushEvery)
	defer t.Stop()

	readers := make(chan io.Reader)

	go func() {
		defer close(readers)

		pr, pw := io.Pipe()
		readers <- pr

		scanner := bufio.NewScanner(r)
		scanner.Split(splitLinesMulti)

		for {
			select {
			case <-ctx.Done():
				_ = pw.CloseWithError(scanner.Err())
				return
			case <-t.C:
				if err := pw.Close(); err != nil {
					return
				}

				pr, pw = io.Pipe()
				readers <- pr
			default:
				if !scanner.Scan() {
					_ = pw.CloseWithError(scanner.Err())
					return
				}

				if _, err := pw.Write(scanner.Bytes()); err != nil {
					_ = pw.CloseWithError(err)
					return
				}
			}
		}
	}()

	res := new(axiom.IngestStatus)
	for r := range readers {
		ingestRes, err := ingest(ctx, client, r, axiom.NDJSON, opts)
		if err != nil {
			return res, err
		}
		mergeIngestStatuses(res, ingestRes)
	}

	return res, nil
}

func ingest(ctx context.Context, client *axiom.Client, r io.Reader, typ axiom.ContentType, opts *options) (*axiom.IngestStatus, error) {
	enc := axiom.Identity
	if opts.Compression {
		var err error
		if r, err = axiom.GZIPStreamer(r, gzip.BestSpeed); err != nil {
			return nil, fmt.Errorf("could not apply compression: %w", err)
		}
		enc = axiom.GZIP
	}

	res, err := client.Datasets.Ingest(ctx, opts.Dataset, r, typ, enc, axiom.IngestOptions{
		TimestampField:  opts.TimestampField,
		TimestampFormat: opts.TimestampFormat,
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

// detectContentType detects the content type of an io.Reader's data. It returns
// a new reader which must be used in place of the old one.
func detectContentType(r io.Reader) (io.Reader, axiom.ContentType, error) {
	var (
		br  = bufio.NewReader(r)
		typ axiom.ContentType
	)
	for {
		var (
			c   rune
			err error
		)
		if c, _, err = br.ReadRune(); err == io.EOF {
			return nil, 0, errors.New("couldn't find beginning of valid JSON")
		} else if err != nil {
			return nil, 0, err
		} else if c == '[' {
			typ = axiom.JSON
		} else if c == '{' {
			typ = axiom.NDJSON
		} else if unicode.IsSpace(c) {
			continue
		} else {
			return nil, 0, errors.New("cannot determine content type")
		}

		if err = br.UnreadRune(); err != nil {
			return nil, 0, err
		}
		break
	}

	// Create a new reader and prepend what we have already consumed in order to
	// figure out the content type.
	bufSize := br.Buffered()
	buf, err := br.Peek(bufSize)
	if err != nil {
		return nil, 0, err
	}
	alreadyRead := bytes.NewReader(buf)
	r = io.MultiReader(alreadyRead, r)

	return r, typ, nil
}

func mergeIngestStatuses(base, add *axiom.IngestStatus) {
	if base == nil || add == nil {
		return
	}

	base.Ingested += add.Ingested
	base.Failed += add.Failed
	base.Failures = append(base.Failures, add.Failures...)
	base.ProcessedBytes += add.ProcessedBytes
	base.BlocksCreated += add.BlocksCreated
	base.WALLength += add.WALLength
}

// splitLinesMulti is like bufio.SplitLines, but returns multiple lines
// including the newline char.
func splitLinesMulti(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.LastIndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, data[0 : i+1], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}
