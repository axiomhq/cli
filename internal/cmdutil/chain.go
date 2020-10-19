package cmdutil

import (
	"text/template"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/pkg/terminal"
)

var (
	noDeploymentsMsg = heredoc.Doc(`
		{{ errorIcon }} No deployments configured!

		  Setup a deployment by logging into it:
		  $ {{ bold "axiom auth login" }}
	`)

	badDeploymentMsg = heredoc.Doc(`
		{{ errorIcon }} Chosen deployment {{ bold .Deployment }} is not configured!
	`)

	badActiveDeploymentMsg = heredoc.Doc(`
		{{ errorIcon }} Chosen deployment {{ bold .Deployment }} is not configured!

		  Select a deployment which is persisted by setting the {{ bold "active_deployment" }}
		  key in the configuration file currently in use:
		  $ {{ bold "axiom auth select" }}
		  
		  Select a deployment by setting the {{ bold "AXM_DEPLOYMENT" }} environment variable. This
		  overwrittes the choice made in the configuration file: 
		  $ {{ bold "export AXM_DEPLOYMENT=axiom-eu-west-1" }}

		  Select a deployment by setting the {{ bold "-D" }} or {{ bold "--deployment" }} flag. This
		  overwrittes the choice made in the configuration file or the environment: 
		  $ {{ bold .CommandPath }} {{ bold "-D=axiom-eu-west-1" }}
	`)

	noDatasetsMsg = heredoc.Doc(`
		{{ errorIcon }} No datasets present on configured deployment {{ bold .Deployment }}!

		  Explicitly create a datatset on the configured deployment:
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

// NeedsActiveDeployment makes sure an active deployment is configured. If not,
// it asks for one when the application is running interactively. If no
// deployments are configured or the application is not running interactively,
// an error is printed and a silent error returned.
func NeedsActiveDeployment(f *Factory) RunFunc {
	return func(cmd *cobra.Command, _ []string) error {
		// If no deployments are configured, print an error message.
		if len(f.Config.Deployments) == 0 {
			return execTemplateSilent(f.IO, noDeploymentsMsg, nil)
		}

		// If the given deployment is configured, that's all we need. If it is
		// not configured, but given, print an error message.
		if _, ok := f.Config.GetActiveDeployment(); ok {
			return nil
		} else if f.Config.ActiveDeployment != "" {
			return execTemplateSilent(f.IO, badActiveDeploymentMsg, map[string]string{
				"Deployment":  f.Config.ActiveDeployment,
				"CommandPath": cmd.CommandPath(),
			})
		}

		// When not running interactively and no active deployment is given, the
		// deployment to use must be provided as a flag.
		if !f.IO.IsStdinTTY() && f.Config.ActiveDeployment == "" {
			return NewFlagErrorf("--deployment or -D required when not running interactively")
		}

		return survey.AskOne(&survey.Select{
			Message: "Which deployment to use?",
			Options: f.Config.DeploymentAliases(),
		}, &f.Config.ActiveDeployment, f.IO.SurveyIO())
	}
}

// NeedsDeployments prints an error message and errors silently if no
// deployments are configured.
func NeedsDeployments(f *Factory) RunFunc {
	return func(cmd *cobra.Command, _ []string) error {
		if len(f.Config.Deployments) == 0 {
			return execTemplateSilent(f.IO, noDeploymentsMsg, nil)
		}
		return nil
	}
}

// NeedsValidDeployment prints an error message and errors silently if the given
// deployment is not configured. An empty alias is not evaluated.
func NeedsValidDeployment(f *Factory, alias *string) RunFunc {
	return func(cmd *cobra.Command, _ []string) error {
		if _, ok := f.Config.Deployments[*alias]; !ok && *alias != "" {
			return execTemplateSilent(f.IO, badDeploymentMsg, map[string]string{
				"Deployment": *alias,
			})
		}
		return nil
	}
}

// NeedsDatasets prints an error message and errors silently if no datasets are
// available on the configured deployment.
func NeedsDatasets(f *Factory) RunFunc {
	return func(cmd *cobra.Command, _ []string) error {
		client, err := f.Client()
		if err != nil {
			return err
		}

		datasets, err := client.Datasets.List(cmd.Context())
		if err != nil {
			return err
		}

		if len(datasets) == 0 {
			return execTemplateSilent(f.IO, noDatasetsMsg, map[string]string{
				"Deployment": f.Config.ActiveDeployment,
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
