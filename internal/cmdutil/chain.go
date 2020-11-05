package cmdutil

import (
	"text/template"

	axiomdb "axicode.axiom.co/watchmakers/axiomdb/client"
	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/pkg/terminal"
)

var (
	noBackendsMsg = heredoc.Doc(`
		{{ errorIcon }} No backends configured!

		  Setup a backend by logging into it:
		  $ {{ bold "axiom auth login" }}
	`)

	badBackendMsg = heredoc.Doc(`
		{{ errorIcon }} Chosen backend {{ bold .Backend }} is not configured!
	`)

	badActiveBackendMsg = heredoc.Doc(`
		{{ errorIcon }} Chosen backend {{ bold .Backend }} is not configured!

		  Select a backend which is persisted by setting the {{ bold "active_backend" }}
		  key in the configuration file currently in use:
		  $ {{ bold "axiom auth select" }}
		  
		  Select a backend by setting the {{ bold "AXM_BACKEND" }} environment variable. This
		  overwrittes the choice made in the configuration file: 
		  $ {{ bold "export AXM_BACKEND=my-axiom" }}

		  Select a backend by setting the {{ bold "-B" }} or {{ bold "--backend" }} flag. This overwrittes
		  the choice made in the configuration file or the environment: 
		  $ {{ bold .CommandPath }} {{ bold "-B=my-axiom" }}
	`)

	noDatasetsMsg = heredoc.Doc(`
		{{ errorIcon }} No datasets present on configured backend {{ bold .Backend }}!

		  Explicitly create a datatset on the configured backend:
		  $ {{ bold "axiom dataset create" }}
		  $ {{ bold "axiom dataset create nginx-logs" }}

		  Have the dataset created as part of ingestion into a named dataset:
		  $ {{ bold "cat logs.json | axiom ingest -d create" }}
	`)
)

// RunFunc is a cobra run function which is compatible with PersistentPreRunE,
// PreRunE, RunE, PostRunE and PersistentPostRunE.
type RunFunc func(cmd *cobra.Command, args []string) error

// ChainRunFuncs chains one or more RunFunc's.
func ChainRunFuncs(fns ...RunFunc) RunFunc {
	return func(cmd *cobra.Command, args []string) error {
		for _, fn := range fns {
			if err := fn(cmd, args); err != nil {
				return err
			}
		}
		return nil
	}
}

// NeedsRootPersistentPreRunE executes the root commands PersistentPreRunE
// function.
func NeedsRootPersistentPreRunE() RunFunc {
	return func(cmd *cobra.Command, args []string) error {
		return cmd.Root().PersistentPreRunE(cmd, args)
	}
}

// NeedsActiveBackend makes sure an active backend is configured. If not, it
// asks for one when the application is running interactively. If no backends to
// select from are configured or the application is not running interactively,
// an error is printed and a silent error returned.
func NeedsActiveBackend(f *Factory) RunFunc {
	return func(cmd *cobra.Command, _ []string) error {
		// If no backends are configured, print an error message.
		if len(f.Config.Backends) == 0 {
			return execTemplateSilent(f.IO, noBackendsMsg, nil)
		}

		// If the given backend is configured, that's all we need. If it is not
		// configured, but given, print an error message.
		if _, ok := f.Config.Backends[f.Config.ActiveBackend]; ok {
			return nil
		} else if f.Config.ActiveBackend != "" {
			return execTemplateSilent(f.IO, badActiveBackendMsg, map[string]string{
				"Backend":     f.Config.ActiveBackend,
				"CommandPath": cmd.CommandPath(),
			})
		}

		// When not running interactively and no active backend is given, the
		// backend to use must be provided as a flag.
		if !f.IO.IsStdinTTY() && f.Config.ActiveBackend == "" {
			return NewFlagErrorf("--backend or -B required when not running interactively")
		}

		return survey.AskOne(&survey.Select{
			Message: "Which backend to use?",
			Options: f.Config.BackendAliases(),
		}, &f.Config.ActiveBackend, f.IO.SurveyIO())
	}
}

// NeedsBackends prints an error message and errors silently if no backends are
// configured.
func NeedsBackends(f *Factory) RunFunc {
	return func(cmd *cobra.Command, _ []string) error {
		if len(f.Config.Backends) == 0 {
			return execTemplateSilent(f.IO, noBackendsMsg, nil)
		}
		return nil
	}
}

// NeedsValidBackend prints an error message and errors silently if the given
// backend is not configured. An empty alias is not evaluated.
func NeedsValidBackend(f *Factory, alias *string) RunFunc {
	return func(cmd *cobra.Command, _ []string) error {
		if _, ok := f.Config.Backends[*alias]; !ok && *alias != "" {
			return execTemplateSilent(f.IO, badBackendMsg, map[string]string{
				"Backend": *alias,
			})
		}
		return nil
	}
}

// NeedsDatasets prints an error message and errors silently if no datasets are
// available on the configured backend.
func NeedsDatasets(f *Factory) RunFunc {
	return func(cmd *cobra.Command, _ []string) error {
		client, err := f.Client()
		if err != nil {
			return err
		}

		datasets, err := client.Datasets.List(cmd.Context(), axiomdb.ListOptions{})
		if err != nil {
			return err
		}

		if len(datasets) == 0 {
			return execTemplateSilent(f.IO, noDatasetsMsg, map[string]string{
				"Backend": f.Config.ActiveBackend,
			})
		}

		return nil
	}
}

// execTemplateSilent parses and executes a template, but still returns
// ErrSilent on success.
func execTemplateSilent(io *terminal.IO, tmplStr string, data interface{}) (err error) {
	tmpl := template.New("util").Funcs(io.ColorScheme().TemplateFuncs())
	if tmpl, err = tmpl.Parse(tmplStr); err != nil {
		return err
	} else if err = tmpl.Execute(io.ErrOut(), data); err != nil {
		return err
	}
	return ErrSilent
}
