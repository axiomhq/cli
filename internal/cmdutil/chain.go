package cmdutil

import (
	"text/template"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/client"
	"github.com/axiomhq/cli/internal/config"
	"github.com/axiomhq/cli/pkg/surveyext"
	"github.com/axiomhq/cli/pkg/terminal"
)

var (
	noDeploymentsMsgTmpl = heredoc.Doc(`
		{{ errorIcon }} No deployments configured!

		  Setup a deployment by logging into it:
		  $ {{ bold "axiom auth login" }}
	`)

	badDeploymentMsgTmpl = heredoc.Doc(`
		{{ errorIcon }} Chosen deployment {{ bold .Deployment }} is not configured!
	`)

	badActiveDeploymentMsgTmpl = heredoc.Doc(`
		{{ errorIcon }} Chosen deployment {{ bold .Deployment }} is not configured!

		  Select a deployment which is persisted by setting the {{ bold "active_deployment" }}
		  key in the configuration file currently in use:
		  $ {{ bold "axiom auth select" }}
		  
		  Select a deployment by setting the {{ bold "AXIOM_DEPLOYMENT" }} environment variable. This
		  overwrittes the choice made in the configuration file: 
		  $ {{ bold "export AXIOM_DEPLOYMENT=axiom-eu-west-1" }}

		  Select a deployment by setting the {{ bold "-D" }} or {{ bold "--deployment" }} flag. This
		  overwrittes the choice made in the configuration file or the environment: 
		  $ {{ bold .CommandPath }} {{ bold "-D=axiom-eu-west-1" }}

		  For non-interactive use, set AXIOM_TOKEN and AXIOM_URL to target a deployment directly,
		  without first configuring it.
	`)

	intialSetupSkippedMsgTmpl = heredoc.Doc(`
		{{ warningIcon }} Skipped setup. Most functionality will be limited.

		To login to Axiom, run:
		$ {{ bold "axiom auth login" }}
	`)

	noDatasetsMsgTmpl = heredoc.Doc(`
		{{ errorIcon }} No datasets present on configured deployment!

		  Explicitly create a dataset on the configured deployment:
		  $ {{ bold "axiom dataset create" }}
		  $ {{ bold "axiom dataset create http-logs" }}
	`)

	noPersonalAccessTokenGiven = heredoc.Doc(`
		{{ errorIcon }} Deployment is not configured with a personal access token!

		  To run {{ bold .CommandPath }} make sure to use a Personal Access token.
		  Help on tokens:
		  $ {{ bold "axiom help credentials" }}

		  To update the token for the deployment, run:
		  $ {{ bold "axiom auth update-token" }}
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

// AsksForSetup will ask the user to setup Axiom in case it is not yet
// configured.
func AsksForSetup(f *Factory, loginCmd *cobra.Command) RunFunc {
	return func(cmd *cobra.Command, _ []string) error {
		if config.HasDefaultConfigFile() || !f.Config.IsEmpty() || !f.IO.IsStdinTTY() || !f.IO.IsStdoutTTY() || !f.IO.IsStderrTTY() {
			return nil
		}

		if ok, err := surveyext.AskConfirm("This seems to be your first time running this CLI. Do you want to login to Axiom?", true, f.IO.SurveyIO()); err != nil {
			return err
		} else if !ok {
			// Write default config file to prevent this message from showing up
			// again.
			if err := f.Config.Write(); err != nil {
				return err
			}
			return execTemplateSilent(f.IO, intialSetupSkippedMsgTmpl, nil)
		}

		return loginCmd.ExecuteContext(cmd.Context())
	}
}

// NeedsActiveDeployment makes sure an active deployment is configured. If not,
// it asks for one when the application is running interactively. If no
// deployments are configured or the application is not running interactively,
// an error is printed and a silent error returned.
func NeedsActiveDeployment(f *Factory) RunFunc {
	return func(cmd *cobra.Command, _ []string) error {
		// If the given deployment is configured, that's all we need. If it is
		// not configured, but given, print an error message.
		if _, ok := f.Config.GetActiveDeployment(); ok {
			return nil
		} else if f.Config.ActiveDeployment != "" {
			return execTemplateSilent(f.IO, badActiveDeploymentMsgTmpl, map[string]string{
				"Deployment":  f.Config.ActiveDeployment,
				"CommandPath": cmd.CommandPath(),
			})
		}

		// At this point, we need at least one configured deployment.
		if len(f.Config.Deployments) == 0 {
			return execTemplateSilent(f.IO, noDeploymentsMsgTmpl, nil)
		}

		// When not running interactively and no active deployment is given, the
		// deployment to use must be provided as a flag.
		if !f.IO.IsStdinTTY() && f.Config.ActiveDeployment == "" {
			return NewFlagErrorf("--deployment or -D required when not running interactively")
		}

		options := f.Config.DeploymentAliases()
		return survey.AskOne(&survey.Select{
			Message: "Which deployment to use?",
			Default: options[0],
			Options: options,
		}, &f.Config.ActiveDeployment, f.IO.SurveyIO())
	}
}

// NeedsDeployments prints an error message and errors silently if no
// deployments are configured.
func NeedsDeployments(f *Factory) RunFunc {
	return func(cmd *cobra.Command, _ []string) error {
		if len(f.Config.Deployments) == 0 {
			return execTemplateSilent(f.IO, noDeploymentsMsgTmpl, nil)
		}
		return nil
	}
}

// NeedsValidDeployment prints an error message and errors silently if the given
// deployment is not configured. An empty alias is not evaluated.
func NeedsValidDeployment(f *Factory, alias *string) RunFunc {
	return func(cmd *cobra.Command, _ []string) error {
		if _, ok := f.Config.Deployments[*alias]; !ok && *alias != "" {
			return execTemplateSilent(f.IO, badDeploymentMsgTmpl, map[string]string{
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
		// Skip if token is not a Personal Access Token.
		if dep, ok := f.Config.GetActiveDeployment(); ok && !client.IsPersonalToken(dep.Token) {
			return nil
		}

		client, err := f.Client(cmd.Context())
		if err != nil {
			return err
		}

		datasets, err := client.Datasets.List(cmd.Context())
		if err != nil {
			return err
		}

		if len(datasets) == 0 {
			return execTemplateSilent(f.IO, noDatasetsMsgTmpl, map[string]string{
				"Deployment": f.Config.ActiveDeployment,
			})
		}

		return nil
	}
}

// NeedsPersonalAccessToken prints an error message and errors silently if the
// active deployment is not configured with a personal access token.
func NeedsPersonalAccessToken(f *Factory) RunFunc {
	return func(cmd *cobra.Command, _ []string) error {
		// We need an active deployment.
		dep, ok := f.Config.GetActiveDeployment()
		if !ok {
			return nil
		}

		if client.IsPersonalToken(dep.Token) {
			return nil
		}

		err := execTemplateSilent(f.IO, noPersonalAccessTokenGiven, map[string]string{
			"Deployment":  f.Config.ActiveDeployment,
			"CommandPath": cmd.CommandPath(),
		})

		return err
	}
}

// execTemplateSilent parses and executes a template, but still returns
// ErrSilent on success.
func execTemplateSilent(io *terminal.IO, tmplStr string, data map[string]string) (err error) {
	tmpl := template.New("util").Funcs(io.ColorScheme().TemplateFuncs())
	if tmpl, err = tmpl.Parse(tmplStr); err != nil {
		return err
	} else if err = tmpl.Execute(io.ErrOut(), data); err != nil {
		return err
	}
	return ErrSilent
}
