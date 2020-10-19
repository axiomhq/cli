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
	"github.com/mgutz/ansi"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/cmd/axiom/root"
	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/internal/config"
	"github.com/axiomhq/cli/pkg/terminal"
)

func main() {
	// Setup signal handling.
	term := make(chan os.Signal, 1)
	defer close(term)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(term)

	// Setup an application lifecycle context.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup the factory which flows through the command call stack.
	f := cmdutil.NewFactory()

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

	// Enable running gh from Windows File Explorer's address bar. Without this,
	// the user is told to stop and run from a terminal.
	// https://github.com/cli/cli/blob/trunk/cmd/gh/main.go#L70
	if len(os.Args) > 1 && os.Args[1] != "" {
		cobra.MousetrapHelpText = ""
	}

	// Add template functions of the color scheme.
	cobra.AddTemplateFuncs(f.IO.ColorScheme().TemplateFuncs())

	// Set pager command, if explicitly set.
	if pager, ok := os.LookupEnv("AXM_PAGER"); ok {
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

	// Cancel the application lifecycle context when receiving a signal.
	go func() {
		select {
		case <-ctx.Done():
		case <-term:
			fmt.Fprintln(f.IO.ErrOut(), "Received interrupt")
			cancel()
		}
	}()

	rootCmd := root.NewRootCmd(f)

	cmdutil.DefaultCompletion(rootCmd)
	cmdutil.InheritRootPersistenPreRun(rootCmd)

	// Finally execute the root command.
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		printError(f.IO.ErrOut(), err, rootCmd)
		os.Exit(1)
	} else if root.HasFailed() {
		os.Exit(1)
	}
}

func printError(w io.Writer, err error, cmd *cobra.Command) {
	// We don't want to print an error if it is explicitly marked as silent or
	// a survey prompt is terminated by interrupt.
	if errors.Is(err, cmdutil.ErrSilent) || errors.Is(err, surveyTerm.InterruptErr) {
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
