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
	"text/template"

	surveyCore "github.com/AlecAivazis/survey/v2/core"
	surveyTerm "github.com/AlecAivazis/survey/v2/terminal"
	"github.com/MakeNowJust/heredoc"
	"github.com/axiomhq/axiom-go/axiom"
	"github.com/mgutz/ansi"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/internal/config"
	"github.com/axiomhq/cli/pkg/surveyext"
	"github.com/axiomhq/cli/pkg/terminal"

	// Commands

	"github.com/axiomhq/cli/internal/cmd/root"
)

var intialSetupSkippedMsgTmpl = heredoc.Doc(`
	{{ warningIcon }} Skipped setup. Most functionality will be limited.

	  To login to Axiom, run:
	  $ {{ bold "axiom login" }}
`)

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

	// Go through setup if the default configuration file is not present, no
	// configuration value is set and a TTY is attached.
	if !config.HasDefaultConfigFile() && f.Config.IsEmpty() && f.IO.IsStdinTTY() {
		// If the user is trying to login or get the version, we don't want to
		// interfer with that.
		if args, l := os.Args[1:], len(os.Args[1:]); l < 1 || l > 2 ||
			l == 2 && (args[0] != "auth" || args[1] != "login") ||
			l == 1 && args[0] != "-v" {
			if ok, err := surveyext.AskConfirm("This seems to be your first time running this CLI. Do you want to login to Axiom?", true, f.IO.SurveyIO()); err != nil {
				printError(f.IO.ErrOut(), err, nil)
				os.Exit(1)
			} else if !ok {
				tmpl := template.New("setup").Funcs(f.IO.ColorScheme().TemplateFuncs())
				if tmpl, err = tmpl.Parse(intialSetupSkippedMsgTmpl); err != nil {
					printError(f.IO.ErrOut(), err, nil)
				} else if err = tmpl.Execute(f.IO.ErrOut(), nil); err != nil {
					printError(f.IO.ErrOut(), err, nil)
				}

				// Write default config file to prevent this message from
				// showing up again.
				_ = f.Config.Write()

				os.Exit(0)
			}

			rootCmd.SetArgs([]string{"auth", "login"})
			if cmd, err := rootCmd.ExecuteContextC(ctx); err != nil {
				printError(f.IO.ErrOut(), err, cmd)
				os.Exit(1)
			} else if root.HasFailed() {
				os.Exit(1)
			}
		}
	}

	// Finally execute the root command.
	rootCmd.SetArgs(os.Args[1:])
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
