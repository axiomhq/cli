package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	surveyCore "github.com/AlecAivazis/survey/v2/core"
	surveyTerm "github.com/AlecAivazis/survey/v2/terminal"
	"github.com/axiomhq/axiom-go/axiom"
	"github.com/mgutz/ansi"
	"github.com/muesli/termenv"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/internal/config"
	"github.com/axiomhq/cli/pkg/terminal"

	// Commands
	"github.com/axiomhq/cli/internal/cmd/root"
)

func main() {
	// Setup signal handling.
	ctx, cancel := signal.NotifyContext(context.Background(),
		os.Interrupt,
		os.Kill,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGINT,
		syscall.SIGHUP,
	)
	defer cancel()

	// Setup the factory which flows through the command call stack.
	f := cmdutil.NewFactory()

	// Setup the I/O for the "browser" package, globally.
	browser.Stdout = f.IO.Out()
	browser.Stderr = f.IO.ErrOut()

	// Set survey colored output, if enabled, and override its poor choice of
	// color.
	if !f.IO.ColorEnabled() {
		surveyCore.DisableColor = true
	} else {
		surveyCore.TemplateFuncsWithColor["color"] = func(style string) string {
			if style == "white" {
				color := terminal.Gray.Color(f.IO.ColorScheme().IsDark())
				return fmt.Sprintf("%s%sm", termenv.CSI, f.IO.ColorScheme().Code(color))
			}

			// TODO(lukasmalkmus): Find a better way using termenv.
			return ansi.ColorCode(style)
		}
	}

	// Enable running axiom from Windows File Explorer's address bar. Without
	// this, the user is told to stop and run from a terminal.
	// Reference: https://github.com/cli/cli/blob/trunk/cmd/gh/main.go#L70
	if len(os.Args) > 1 && os.Args[1] != "" {
		cobra.MousetrapHelpText = ""
	}

	// Add template functions of the color scheme.
	cobra.AddTemplateFuncs(f.IO.ColorScheme().TemplateFuncs())

	// Set pager command, if explicitly set.
	if pager, ok := os.LookupEnv("AXIOM_PAGER"); ok {
		f.IO.SetPagerCommand(pager)
	}

	// Initially load configuration to have it available in completion. However,
	// the config and deployment override flags are only parsed when running
	// commands. This makes completion only work for the configured deployments
	// and not the overwritten ones.
	var err error
	if f.Config, err = config.LoadDefault(); err != nil {
		printError(f.IO.ErrOut(), err, nil)
		os.Exit(2)
	}

	rootCmd := root.NewCmd(f)
	cmdutil.DefaultCompletion(rootCmd)
	cmdutil.InheritRootPersistenPreRun(rootCmd)

	// Finally execute the root command.
	if cmd, err := rootCmd.ExecuteContextC(ctx); err != nil {
		printError(f.IO.ErrOut(), err, cmd)
		os.Exit(1)
	} else if root.HasFailed() {
		os.Exit(1)
	}
}

func printError(w io.Writer, err error, cmd *cobra.Command) {
	// We don't want to print an error if it is explicitly marked as silent or
	// a survey prompt is terminated by interrupt.
	var pagerPipeError *terminal.ErrClosedPagerPipe
	if errors.Is(err, cmdutil.ErrSilent) ||
		errors.Is(err, surveyTerm.InterruptErr) ||
		errors.As(err, &pagerPipeError) {
		return
	}

	// Print some nicer output for Axiom API errors.
	if errors.Is(err, axiom.ErrNotFound) || errors.Is(err, axiom.ErrExists) ||
		errors.Is(err, axiom.ErrUnauthorized) || errors.Is(err, axiom.ErrUnauthenticated) {
		fmt.Fprintf(w, "Error: %s\n", errors.Unwrap(err))
		return
	}

	// Print some nicer errors for context related errors.
	if errors.Is(err, context.Canceled) {
		fmt.Fprintln(w, "Error: Operation was canceled")
		return
	} else if errors.Is(err, context.DeadlineExceeded) {
		fmt.Fprintln(w, "Error: Operation timed out")
		return
	}

	fmt.Fprintf(w, "Error: %s\n", err)

	// Only print the command usage if the error is related to bad user input.
	var flagError *cmdutil.FlagError
	if errors.As(err, &flagError) || strings.HasPrefix(err.Error(), "unknown command ") {
		if !strings.HasSuffix(err.Error(), "\n") {
			fmt.Fprintln(w)
		}

		usage := cmd.UsageString()
		fmt.Fprint(w, usage)
		if !strings.HasSuffix(usage, "\n") {
			fmt.Fprintln(w)
		}
	}
}
