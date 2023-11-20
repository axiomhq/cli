package ingest

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/axiomhq/axiom-go/axiom"
	"github.com/axiomhq/axiom-go/axiom/ingest"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/client"
	"github.com/axiomhq/cli/internal/cmd/auth"
	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/pkg/utils"
)

var (
	validContentTypes = []string{
		axiom.JSON.String(),
		axiom.NDJSON.String(),
		axiom.CSV.String(),
	}

	validContentEncodings = []string{
		axiom.Identity.String(),
		axiom.Gzip.String(),
		axiom.Zstd.String(),
	}
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
	// Delimiter that separates CSV fields.
	Delimiter string
	// FlushEvery flushes the ingestion buffer after the specified duration. It
	// is only valid when ingesting a stream of newline delimited JSON objects
	// of unknown length.
	FlushEvery time.Duration
	// ContentType of the data to ingest.
	ContentType axiom.ContentType
	contentType string // for the flag value
	// ContentEncoding of the data to ingest.
	ContentEncoding axiom.ContentEncoding
	contentEncoding string // for the flag value
	// Labels attached to every event, server-side.
	Labels []ingest.Option
	labels []string // for the flag value
	// ContinueOnError will continue ingesting, even if an error is returned
	// from the server.
	ContinueOnError bool
}

// NewCmd creates and returns the ingest command.
func NewCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &options{
		Factory: f,
	}

	cmd := &cobra.Command{
		Use:   "ingest <dataset-name> [(-f|--file) <filename> [ ...]] [--timestamp-field <timestamp-field>] [--timestamp-format <timestamp-format>] [--flush-every <duration>] [(-t|--content-type <content-type>] [(-e|--content-encoding <content-encoding>] [(-l|--label) <key>:<value> [ ...]]",
		Short: "Ingest structured data",
		Long: heredoc.Doc(`
			Ingest structured data into an Axiom dataset.

			Supported formats are: Newline delimited JSON (NDJSON), an array of
			JSON objects (JSON) and a newline delimited list of comma separated
			values (CSV). The first line of CSV content is assumed to be the
			field names for the values in the following lines. The input format
			is automatically detected.

			Each object is assigned an event timestamp from the configured
			timestamp field (default "_time"). If there is no timestamp field
			Axiom will assign the server side time of reception. The timestamp
			format can be configured by specifying a pattern with the reference
			date:

				Mon Jan 2 15:04:05 -0700 MST 2006

			Omitted elements in the pattern are treated as zero or one as
			applicable. See the Go reference documentation for examples:
			https://pkg.go.dev/time#pkg-constants

			For Unix timestamps, leave the timestamp format unspecified and just
			provide the value as a number. Can be seconds, milliseconds,
			microseconds or nanoseconds.
		`),

		DisableFlagsInUseLine: true,

		Args:              cmdutil.PopulateFromArgs(f, &opts.Dataset),
		ValidArgsFunction: cmdutil.DatasetCompletionFunc(f),

		Example: heredoc.Doc(`
			# Ingest the contents of a JSON logfile into a dataset named
			# "http-logs":
			$ axiom ingest http-logs -f http-logs.json

			# Pipe the contents of a log generator into a dataset named
			# "gen-logs". If the length of the data stream is unknown, the
			# "--flush-every" flag can be tweaked to optimize shipping the data
			# to the server after the specified duration. This is only valid for
			# newline delimited JSON.
			$ ./loggen -ndjson | axiom ingest gen-logs

			# Send a set of gzip compressed JSON logs to a dataset called
			# "http-logs". The content type is automatically detected. 
			$ zcat log*.json.gz | axiom ingest http-logs
			
			# Send a set of gzip compressed JSON logs to a dataset called
			# "http-logs":
			$ cat log*.json.gz | axiom ingest http-logs -t=json -e=gzip
			
			# Send a set of gzip compressed JSON logs to a dataset called
			# "http-logs" and attach some labels. Labels are added server-side
			# to every events, so there is no need to add them to the data,
			# locally:
			$ cat log*.json.gz | axiom ingest http-logs -t=json -e=gzip -l=env:prod -l=app:webserver
		`),

		Annotations: map[string]string{
			"IsCore": "true",
		},

		PreRunE: cmdutil.ChainRunFuncs(
			cmdutil.AsksForSetup(f, auth.NewLoginCmd(f)),
			cmdutil.NeedsActiveDeployment(f),
			cmdutil.NeedsDatasets(f),
		),

		RunE: func(cmd *cobra.Command, args []string) (err error) {
			// When no files are specified, stdin is the file to use.
			if len(opts.Filenames) == 0 {
				opts.Filenames = []string{"-"}
			}

			// If set, parse content type and content encoding from their string
			// representation.
			if opts.ContentType, err = contentTypeFromString(opts.contentType); err != nil && cmd.Flag("content-type").Changed {
				return err
			} else if opts.ContentEncoding, err = contentEncodingFromString(opts.contentEncoding); err != nil && cmd.Flag("content-encoding").Changed {
				return err
			}

			// If the content encoding is set to anything else than "identity",
			// make sure the content type is set, as well.
			if opts.ContentEncoding != axiom.Identity && opts.ContentType == 0 {
				return fmt.Errorf("content encoding set but content type not set")
			}

			// Sanity check the labels.
			for _, label := range opts.labels {
				splits := strings.Split(label, ":")
				if len(splits) != 2 {
					return fmt.Errorf("malformed label: %q", label)
				}
				opts.Labels = append(opts.Labels, ingest.SetEventLabel(splits[0], splits[1]))
			}

			if err := complete(cmd.Context(), opts); err != nil {
				return err
			}
			return run(cmd.Context(), opts, cmd.Flag("flush-every").Changed)
		},
	}

	cmd.Flags().StringSliceVarP(&opts.Filenames, "file", "f", nil, "File(s) to ingest (- to read from stdin). If stdin is a pipe the default value is -, otherwise this is a required parameter")
	cmd.Flags().StringVar(&opts.TimestampField, "timestamp-field", "", "Field to take the ingestion time from (defaults to _time)")
	cmd.Flags().StringVar(&opts.TimestampFormat, "timestamp-format", "", "Format used in the the timestamp field. Default uses a heuristic parser. Must be expressed using the reference time 'Mon Jan 2 15:04:05 -0700 MST 2006'")
	cmd.Flags().StringVarP(&opts.Delimiter, "delimiter", "d", "", "Delimiter that separates CSV fields (only valid when input is CSV")
	cmd.Flags().DurationVar(&opts.FlushEvery, "flush-every", time.Second, "Buffer flush interval for newline delimited JSON streams of unknown length")
	cmd.Flags().StringVarP(&opts.contentType, "content-type", "t", "", "Content type of the data to ingest (will auto-detect if not set, must be set if content encoding is set and content type is not identity)")
	cmd.Flags().StringVarP(&opts.contentEncoding, "content-encoding", "e", axiom.Identity.String(), "Content encoding of the data to ingest")
	cmd.Flags().StringSliceVarP(&opts.labels, "label", "l", nil, "Labels to attach to the ingested events, server side")
	cmd.Flags().BoolVar(&opts.ContinueOnError, "continue-on-error", false, "Don't fail on ingest errors (use with care!)")

	_ = cmd.RegisterFlagCompletionFunc("timestamp-field", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("timestamp-format", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("delimiter", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("flush-every", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("content-type", contentTypeCompletion)
	_ = cmd.RegisterFlagCompletionFunc("content-encoding", contentEncodingCompletion)
	_ = cmd.RegisterFlagCompletionFunc("label", cmdutil.NoCompletion)
	_ = cmd.RegisterFlagCompletionFunc("continue-on-error", cmdutil.NoCompletion)

	if opts.IO.IsStdinTTY() {
		_ = cmd.MarkFlagRequired("file")
	}

	return cmd
}

func complete(ctx context.Context, opts *options) error {
	if opts.Dataset != "" {
		return nil
	}

	// Just fetch a list of available datasets if a Personal Access Token is
	// used.
	var datasetNames []string
	if dep, ok := opts.Config.GetActiveDeployment(); ok && client.IsPersonalToken(dep.Token) {
		client, err := opts.Client(ctx)
		if err != nil {
			return err
		}

		stop := opts.IO.StartActivityIndicator()
		defer stop()

		datasets, err := client.Datasets.List(ctx)
		if err != nil {
			return err
		}

		stop()

		datasetNames = make([]string, len(datasets))
		for i, dataset := range datasets {
			datasetNames[i] = dataset.Name
		}
	}

	if len(datasetNames) == 0 {
		return errors.New("missing dataset")
	}

	return survey.AskOne(&survey.Select{
		Message: "Which dataset to ingest into?",
		Default: datasetNames[0],
		Options: datasetNames,
	}, &opts.Dataset, opts.IO.SurveyIO())
}

func run(ctx context.Context, opts *options, flushEverySet bool) error {
	client, err := opts.Client(ctx)
	if err != nil {
		return err
	}

	stop := opts.IO.StartActivityIndicator()
	defer stop()

	var (
		res     = new(ingest.Status)
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
		if opts.ContentEncoding == axiom.Identity && opts.ContentType == 0 {
			if r, typ, err = axiom.DetectContentType(rc); err != nil {
				_ = rc.Close()
				lastErr = fmt.Errorf("could not detect %q content type: %w", filename, err)
				break
			}
		} else {
			r = rc
			typ = opts.ContentType
		}

		if flushEverySet && typ != axiom.NDJSON {
			return cmdutil.NewFlagErrorf("--flush-every not valid when content type is not newline delimited JSON")
		}
		if opts.Delimiter != "" && typ != axiom.CSV {
			return cmdutil.NewFlagErrorf("--delimier/-d not valid when content type is not CSV")
		}

		var ingestRes *ingest.Status
		if filename == "stdin" && typ == axiom.NDJSON {
			ingestRes, err = ingestEvery(ctx, client, r, opts)
		} else {
			ingestRes, err = ingestReader(ctx, client, r, typ, opts)
		}

		// Error handling below, so we need to check for nil.
		if ingestRes != nil {
			res.Add(ingestRes)
		}

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
		fmt.Fprintf(opts.IO.ErrOut(), "%s processed\n",
			humanize.Bytes(res.ProcessedBytes),
		)

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
					cs.Gray(fail.Timestamp.Format(time.RFC1123)), fail.Error,
				)
			}
		}
	}

	return lastErr
}

func ingestEvery(ctx context.Context, client *axiom.Client, r io.Reader, opts *options) (*ingest.Status, error) {
	t := time.NewTicker(opts.FlushEvery)
	defer t.Stop()

	readers := make(chan io.Reader)

	go func() {
		defer close(readers)

		// Add first reader.
		pr, pw := io.Pipe()
		readers <- pr

		// Start with a 64 byte buffer, check up until 1 MB per line.
		scanner := bufio.NewScanner(r)
		scanner.Buffer(make([]byte, 64), 1024*1024)
		scanner.Split(splitLinesMulti)

		// We need to scan in a go func to make sure we don't block on
		// `scanner.Scan()`.
		done := make(chan struct{})
		lines := make(chan []byte)
		defer close(lines)
		go func() {
			for {
				select {
				case <-ctx.Done():
					_ = pw.CloseWithError(ctx.Err())
					return
				default:
				}

				if !scanner.Scan() {
					close(done)
					return
				}

				line := make([]byte, len(scanner.Bytes()))
				copy(line, scanner.Bytes())
				lines <- line
			}
		}()

		for {
			select {
			case <-ctx.Done():
				_ = pw.CloseWithError(ctx.Err())
				return
			case <-t.C:
				if err := pw.Close(); err != nil {
					return
				}

				pr, pw = io.Pipe()
				readers <- pr
			case line := <-lines:
				if _, err := pw.Write(line); err != nil {
					_ = pw.CloseWithError(err)
					return
				}
			case <-done:
				_ = pw.Close()
				return
			}
		}
	}()

	var res ingest.Status
	for r := range readers {
		ingestRes, err := ingestReader(ctx, client, r, axiom.NDJSON, opts)
		if err != nil {
			if opts.ContinueOnError {
				fmt.Fprintf(opts.IO.ErrOut(), "%s Failed to ingest: %v, continuing...\n",
					opts.IO.ColorScheme().WarningIcon(), err)
				continue
			}
			return &res, err
		}
		res.Add(ingestRes)
	}

	return &res, nil
}

func ingestReader(ctx context.Context, client *axiom.Client, r io.Reader, typ axiom.ContentType, opts *options) (*ingest.Status, error) {
	// If the data to ingest is not compressed, it gets zstd compressed.
	enc := opts.ContentEncoding
	if enc == axiom.Identity {
		var err error
		if r, err = axiom.ZstdEncoder()(r); err != nil {
			return nil, err
		}
		enc = axiom.Zstd
	} else {
		r = io.NopCloser(r)
	}

	ingestOptions := make([]ingest.Option, 0)
	if v := opts.TimestampField; v != "" {
		ingestOptions = append(ingestOptions, ingest.SetTimestampField(v))
	}
	if v := opts.TimestampFormat; v != "" {
		ingestOptions = append(ingestOptions, ingest.SetTimestampFormat(v))
	}
	if v := opts.Delimiter; v != "" {
		ingestOptions = append(ingestOptions, ingest.SetCSVDelimiter(v))
	}
	ingestOptions = append(ingestOptions, opts.Labels...)

	res, err := client.Datasets.Ingest(ctx, opts.Dataset, r, typ, enc, ingestOptions...)
	if err != nil {
		return nil, err
	}

	return res, nil
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

func contentTypeCompletion(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	res := make([]string, 0, len(validContentTypes))
	for _, contentType := range validContentTypes {
		if strings.HasPrefix(contentType, toComplete) {
			res = append(res, contentType)
		}
	}
	return res, cobra.ShellCompDirectiveNoFileComp
}

func contentEncodingCompletion(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	res := make([]string, 0, len(validContentEncodings))
	for _, contentEncoding := range validContentEncodings {
		if strings.HasPrefix(contentEncoding, toComplete) {
			res = append(res, contentEncoding)
		}
	}
	return res, cobra.ShellCompDirectiveNoFileComp
}

func contentTypeFromString(s string) (ct axiom.ContentType, err error) {
	switch strings.ToLower(s) {
	case "json":
		ct = axiom.JSON
	case "ndjson":
		ct = axiom.NDJSON
	case "csv":
		ct = axiom.CSV
	default:
		err = fmt.Errorf("invalid content type %q", s)
	}
	return ct, err
}

func contentEncodingFromString(s string) (ct axiom.ContentEncoding, err error) {
	switch strings.ToLower(s) {
	case "", "identity":
		ct = axiom.Identity
	case "gzip":
		ct = axiom.Gzip
	case "zstd":
		ct = axiom.Zstd
	default:
		err = fmt.Errorf("invalid content encoding %q", s)
	}
	return ct, err
}
