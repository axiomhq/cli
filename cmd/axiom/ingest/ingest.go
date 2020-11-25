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
	"path/filepath"
	"sync"
	"time"
	"unicode"

	"github.com/MakeNowJust/heredoc"
	"github.com/axiomhq/axiom-go/axiom"
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
	// TimestampFormat the timestamp is formatted with.
	TimestampFormat string
}

// NewIngestCmd creates and returns the ingest command.
func NewIngestCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &options{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "ingest <dataset-name> (-|(-f|--file) <filename>...) [--timestamp-field <timestamp-field>] [--timestamp-format <timestamp-format>]",
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
			$ echo nginx-logs.json | axiom ingest nginx-logs

			# Ingest all files inside /var/logs/nginx with extension ".log" into
			# a dataset named nginx-logs:
			$ axiom ingest nginx-logs -f /var/logs/nginx/*.log
		`),

		Annotations: map[string]string{
			"IsCore": "true",
		},

		PreRunE: cmdutil.NeedsActiveDeployment(f),

		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringSliceVarP(&opts.Filenames, "file", "f", []string{"-"}, "File to ingest")
	cmd.Flags().StringVar(&opts.TimestampField, "timestamp-field", "_time", "Field to take the ingestion time from")
	cmd.Flags().StringVar(&opts.TimestampFormat, "timestamp-format", "", "Format the timestamp is formatted with")

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

	var (
		cs      = opts.IO.ColorScheme()
		res     axiom.IngestStatus
		lastErr error
	)
	for k, filename := range opts.Filenames {
		var (
			rc    io.ReadCloser
			size  int64
			title string
		)
		if filename == "-" {
			rc = opts.IO.In()
			size = -1
			title = cs.Bold("stdin")
		} else {
			var f *os.File
			if f, err = os.Open(filename); err != nil {
				rc.Close()
				lastErr = err
				break
			}
			rc = f

			var stat os.FileInfo
			if stat, err = f.Stat(); err != nil {
				rc.Close()
				lastErr = err
				break
			}

			size = stat.Size()
			title = cs.Bold(filepath.Base(f.Name()))
		}

		if l := len(opts.Filenames); l > 1 {
			prefix := fmt.Sprintf("[%d/%d]", k+1, l)
			title = fmt.Sprintf("%s %s", cs.Green(prefix), title)
		}

		var (
			ingestRes *axiom.IngestStatus
			pbr       = opts.IO.StartProgressIndicator(rc, size, title)
		)
		if ingestRes, err = ingest(ctx, client, pbr, opts); err != nil {
			rc.Close()
			lastErr = fmt.Errorf("could not ingest into dataset %q: %w", opts.Dataset, err)
			break
		}
		rc.Close()

		res.Ingested += ingestRes.Ingested
		res.Failed += ingestRes.Failed
		res.BlocksCreated += ingestRes.BlocksCreated
		res.WALLength += ingestRes.WALLength
		res.Failures = append(res.Failures, ingestRes.Failures...)
	}

	if opts.IO.IsStderrTTY() {
		if res.Ingested > 0 {
			fmt.Fprintf(opts.IO.ErrOut(), "%s Ingested %s\n",
				cs.SuccessIcon(),
				utils.Pluralize(cs, "event", int(res.Ingested)),
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

func ingest(ctx context.Context, client *axiom.Client, r io.Reader, opts *options) (*axiom.IngestStatus, error) {
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
			return nil, errors.New("couldn't find beginning of valid JSON")
		} else if err != nil {
			return nil, err
		} else if c == '[' {
			typ = axiom.JSON
		} else if c == '{' {
			typ = axiom.NDJSON
		} else if unicode.IsSpace(c) {
			continue
		} else {
			return nil, errors.New("cannot determine content type")
		}

		if err = br.UnreadRune(); err != nil {
			return nil, err
		}
		break
	}

	// Create a new reader and prepend what we have already consumed in order to
	// figure out the content type.
	bufSize := br.Buffered()
	buf, err := br.Peek(bufSize)
	if err != nil {
		return nil, err
	}
	alreadyRead := bytes.NewReader(buf)
	r = io.MultiReader(alreadyRead, r)

	// Apply compression and capture the error.
	var (
		compressErr    error
		compressErrMtx = new(sync.Mutex)
	)
	errFunc := func(err error) {
		compressErrMtx.Lock()
		defer compressErrMtx.Unlock()

		compressErr = err
	}
	r = gzipStream(r, errFunc)

	res, err := client.Datasets.Ingest(ctx, opts.Dataset, r, typ, axiom.GZIP, axiom.IngestOptions{
		TimestampField:  opts.TimestampField,
		TimestampFormat: opts.TimestampFormat,
	})
	if err != nil {
		compressErrMtx.Lock()
		defer compressErrMtx.Unlock()

		if compressErr != nil {
			return nil, compressErr
		}
		return nil, err
	}

	return res, nil
}

func gzipStream(r io.Reader, errFunc func(error)) io.Reader {
	pr, pw := io.Pipe()
	go func(r io.Reader) {
		defer pw.Close()

		// Does not fail when using a predefined compression level.
		gzw, _ := gzip.NewWriterLevel(pw, gzip.BestSpeed)
		defer gzw.Close()

		if _, err := io.Copy(gzw, r); err != nil {
			errFunc(fmt.Errorf("error compressing data to ingest: %w", err))
		}
	}(r)

	return pr
}
