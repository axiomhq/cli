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

	"github.com/MakeNowJust/heredoc"
	"github.com/axiomhq/axiom-go/axiom"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/pkg/utils"
)

type options struct {
	*cmdutil.Factory

	// Dataset to query.
	Dataset string
	// Filenames of the files to ingest. If not set, will read from Stdin.
	Filenames []string
	// TimestampField to take the ingestion time from.
	TimestampField string
	// TimestampFormat the timestamp is formatted in.
	TimestampFormat string
	// FlushEvery flushes the ingestion buffer after the specified duration. It
	// is only valid when ingesting a stream of newline delimited JSON objects.
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

		Args: cmdutil.ChainPositionalArgs(
			cobra.ExactArgs(1),
			cmdutil.PopulateFromArgs(f, &opts.Dataset),
		),
		ValidArgsFunction: cmdutil.DatasetCompletionFunc(f),

		Example: heredoc.Doc(`
			# Pipe the contents of a JSON logfile into a dataset named
			# "nginx-logs":
			$ cat nginx-logs.json | axiom ingest nginx-logs

			# Ingest all files inside /var/logs/nginx with extension ".log" into
			# a dataset named nginx-logs:
			$ axiom ingest nginx-logs -f /var/logs/nginx/*.log
		`),

		Annotations: map[string]string{
			"IsCore": "true",
		},

		PreRunE: cmdutil.NeedsActiveDeployment(f),

		RunE: func(cmd *cobra.Command, args []string) error {
			if len(opts.Filenames) == 0 {
				opts.Filenames = []string{"-"}
			}
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringSliceVarP(&opts.Filenames, "file", "f", nil, "File to ingest")
	cmd.Flags().StringVar(&opts.TimestampField, "timestamp-field", "", "Field to take the ingestion time from")
	cmd.Flags().StringVar(&opts.TimestampFormat, "timestamp-format", "", "Format the timestamp is formatted in")
	cmd.Flags().DurationVar(&opts.FlushEvery, "flush-every", 5*time.Second, "Buffer flush interval for newline delimited JSON streams")
	cmd.Flags().BoolVarP(&opts.Compression, "compression", "c", true, "Enable gzip compression")

	_ = cmd.RegisterFlagCompletionFunc("timestamp-field", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("timestamp-format", cmdutil.NoCompletion)

	if opts.IO.IsStdinTTY() {
		_ = cmd.MarkFlagRequired("file")
	}

	return cmd
}

func run(ctx context.Context, opts *options) error {
	client, err := opts.Client()
	if err != nil {
		return err
	}

	stop := opts.IO.StartActivityIndicator()

	var (
		res     axiom.IngestStatus
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

		enc := axiom.Identity
		if opts.Compression {
			if r, err = axiom.GZIPStreamer(r, gzip.BestSpeed); err != nil {
				_ = rc.Close()
				lastErr = fmt.Errorf("could not apply compression to %q: %w", filename, err)
				break
			}
			enc = axiom.GZIP
		}

		var ingestRes *axiom.IngestStatus
		if ingestRes, err = ingest(ctx, client, r, typ, enc, opts); err != nil {
			_ = rc.Close()
			lastErr = fmt.Errorf("could not ingest %q into dataset %q: %w", filename, opts.Dataset, err)
			break
		} else if err = rc.Close(); err != nil {
			lastErr = fmt.Errorf("failed to close %q: %w", filename, err)
			break
		}

		res.Ingested += ingestRes.Ingested
		res.Failed += ingestRes.Failed
		res.ProcessedBytes += ingestRes.ProcessedBytes
		res.BlocksCreated += ingestRes.BlocksCreated
		res.WALLength += ingestRes.WALLength
		res.Failures = append(res.Failures, ingestRes.Failures...)
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

func ingest(ctx context.Context, client *axiom.Client, r io.Reader, typ axiom.ContentType, enc axiom.ContentEncoding, opts *options) (*axiom.IngestStatus, error) {
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
