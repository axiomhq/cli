package terminal

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/briandowns/spinner"
	"github.com/cli/safeexec"
	"github.com/google/shlex"
	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
	"golang.org/x/term"

	"github.com/axiomhq/cli/pkg/doc"
)

const (
	defaultTerminalWidth = 80

	spinnerType = 11
)

// ErrClosedPagerPipe is the error returned when writing to a pager that has
// been closed.
type ErrClosedPagerPipe struct {
	error
}

// IO is used to interact with the terminal, files and any other I/O sources.
type IO struct {
	in      io.ReadCloser
	out     io.Writer
	errOut  io.Writer
	origOut io.Writer

	isStdinTTY  bool
	isStdoutTTY bool
	isStderrTTY bool

	colorScheme  *ColorScheme
	colorEnabled bool

	pagerCommand             string
	activityIndicator        *spinner.Spinner
	activityIndicatorEnabled bool
}

// NewIO returns a new IO. It detects if a TTY is attached and detects if
// colored output should be enabled and auto senses the color scheme.
func NewIO() *IO {
	io := &IO{
		in:      os.Stdin,
		out:     colorable.NewColorable(os.Stdout),
		errOut:  colorable.NewColorable(os.Stderr),
		origOut: os.Stdout,

		isStdinTTY:  isTerminal(os.Stdin),
		isStdoutTTY: isTerminal(os.Stdout),
		isStderrTTY: isTerminal(os.Stderr),

		pagerCommand: os.Getenv("PAGER"),
	}

	io.colorEnabled = envColorForced() || (!envColorDisabled() && io.isStdoutTTY)
	io.colorScheme = NewColorScheme(io.colorEnabled)

	if io.isStdoutTTY && io.isStderrTTY {
		color := "fgMagenta"
		if io.colorScheme.dark {
			color = "fgHiMagenta"
		}

		io.activityIndicator = spinner.New(
			spinner.CharSets[spinnerType],
			time.Millisecond*150,
			spinner.WithWriter(io.errOut),
			spinner.WithColor(color),
		)
	}

	return io
}

// EnableActivityIndicator enables or disables the activity indicator. It does
// not force-enable it, if no TTY is attached.
func (io *IO) EnableActivityIndicator(enable bool) {
	io.activityIndicatorEnabled = enable
}

// In returns the input reader.
func (io *IO) In() io.ReadCloser {
	return io.in
}

// Out returns the output writer.
func (io *IO) Out() io.Writer {
	return io.out
}

// ErrOut returns the error output writer.
func (io *IO) ErrOut() io.Writer {
	return io.errOut
}

// ColorEnabled returns true if colored output is enabled.
func (io *IO) ColorEnabled() bool {
	return io.colorEnabled
}

// IsStdinTTY returns true if a TTY is attached to stdin.
func (io *IO) IsStdinTTY() bool {
	return io.isStdinTTY
}

// IsStdoutTTY returns true if a TTY is attached to stdout.
func (io *IO) IsStdoutTTY() bool {
	return io.isStdoutTTY
}

// IsStderrTTY returns true if a TTY is attached to stderr.
func (io *IO) IsStderrTTY() bool {
	return io.isStderrTTY
}

// ColorScheme returns the IO's color scheme used to colorize text output if
// enabled.
func (io *IO) ColorScheme() *ColorScheme {
	return io.colorScheme
}

// SetPagerCommand sets the pager command to use.
func (io *IO) SetPagerCommand(command string) {
	io.pagerCommand = command
}

// StartPager starts the configured pager command and stream IO to it. In case
// the pager command is not set, set to "cat" or no TTY is attached, this is a
// nop.
func (io *IO) StartPager(ctx context.Context) (func(), error) {
	if io.pagerCommand == "" || io.pagerCommand == "cat" || !io.isStdoutTTY {
		return func() {}, nil
	}

	args, err := shlex.Split(io.pagerCommand)
	if err != nil {
		return func() {}, err
	}

	// Get additional pager configuration from environment.
	env := os.Environ()
	for i := len(env) - 1; i >= 0; i-- {
		if strings.HasPrefix(env[i], "PAGER=") {
			env = append(env[0:i], env[i+1:]...)
		}
	}
	if _, ok := os.LookupEnv("LESS"); !ok {
		env = append(env, "LESS=FRX")
	}
	if _, ok := os.LookupEnv("LV"); !ok {
		env = append(env, "LV=-c")
	}

	exe, err := safeexec.LookPath(args[0])
	if err != nil {
		return func() {}, err
	}

	cmd := exec.CommandContext(ctx, exe, args[1:]...)
	cmd.Env = env
	cmd.Stdout = io.out
	cmd.Stderr = io.errOut

	out, err := cmd.StdinPipe()
	if err != nil {
		return func() {}, err
	}
	prevOut := io.out
	io.out = &pagerWriter{out}

	if err = cmd.Start(); err != nil {
		return func() {}, err
	}

	// When called, stops the pager programm and restores the previous output.
	cancelFunc := func() {
		_ = out.Close()
		_, _ = cmd.Process.Wait()
		io.out = prevOut
	}

	return cancelFunc, nil
}

// StartActivityIndicator starts a spinner that indicates activity. The return
// function, when called, stops the spinner. When no TTY is attached, this is a
// nop.
func (io *IO) StartActivityIndicator() func() {
	if !io.isStdoutTTY || !io.isStderrTTY || !io.activityIndicatorEnabled {
		return func() {}
	}
	io.activityIndicator.Start()

	return func() { io.activityIndicator.Stop() }
}

// SurveyIO returns an options that makes itself the IO for survey questions.
// Returns a nop option if no TTY is attached.
func (io *IO) SurveyIO() survey.AskOpt {
	if !io.isStdinTTY || !io.isStdoutTTY {
		return func(*survey.AskOptions) error { return nil }
	}
	return survey.WithStdio(io.in.(*os.File), io.origOut.(*os.File), io.errOut)
}

// TerminalWidth reports the terminals width, if one is attached. If not,
// reports the default of 80 for common 80x24 sized terminals.
func (io *IO) TerminalWidth() int {
	if !io.isStdoutTTY || !io.isStderrTTY {
		return defaultTerminalWidth
	}

	f := io.origOut.(*os.File)
	if w, _, err := term.GetSize(int(f.Fd())); err == nil {
		return w
	}

	if !isatty.IsCygwinTerminal(f.Fd()) {
		return defaultTerminalWidth
	}

	cmd := exec.Command("tput", "cols")
	cmd.Stdin = io.in

	out, err := cmd.Output()
	if err != nil {
		return defaultTerminalWidth
	}

	w, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err == nil {
		return w
	}

	return defaultTerminalWidth
}

// Doc wraps a block of heredoc formatted text at terminal approximately
// terminal width.
func (io *IO) Doc(s string) string {
	return doc.Wrap(s, io.TerminalWidth()-2)
}

func isTerminal(f *os.File) bool {
	fd := f.Fd()
	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}

// TestIO returns an IO which does not read or write to any real outputs.
func TestIO() *IO {
	return &IO{
		in:      ioutil.NopCloser(strings.NewReader("")),
		out:     ioutil.Discard,
		errOut:  ioutil.Discard,
		origOut: ioutil.Discard,
	}
}

// pagerWriter implements an `io.WriteCloser`` that wraps all EPIPE errors in an
// `ErrClosedPagerPipe` type.
type pagerWriter struct {
	io.WriteCloser
}

func (w *pagerWriter) Write(d []byte) (int, error) {
	n, err := w.WriteCloser.Write(d)
	if err != nil && (errors.Is(err, io.ErrClosedPipe) || isEpipeError(err)) {
		return n, &ErrClosedPagerPipe{err}
	}
	return n, err
}
