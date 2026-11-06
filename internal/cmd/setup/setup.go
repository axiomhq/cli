package setup

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmd/auth"
	"github.com/axiomhq/cli/internal/cmdutil"
)

type options struct {
	*cmdutil.Factory

	Agent      string
	Dir        string
	ConfigPath string
}

// NewCmd creates and returns the setup command.
func NewCmd(f *cmdutil.Factory) *cobra.Command {
	opts := &options{Factory: f}

	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Onboard a repository to Axiom with a coding agent",
		Long: heredoc.Doc(`
			Check this repository for an existing OpenTelemetry setup, record which
			Axiom deployment this CLI is logged into, and start Claude Code or Codex
			on the axiom-setup-observability skill.

			The agent instruments the repository. This command does not edit it.
		`),

		Example: heredoc.Doc(`
			$ axiom setup
			$ axiom setup --agent codex
			$ axiom setup --dir ../my-app
		`),

		PreRunE: cmdutil.ChainRunFuncs(
			cmdutil.AsksForSetup(f, auth.NewLoginCmd(f)),
			cmdutil.NeedsActiveDeployment(f),
		),

		RunE: func(cmd *cobra.Command, _ []string) error {
			if flag := cmd.Flag("config"); flag != nil && flag.Changed {
				opts.ConfigPath = flag.Value.String()
			}
			return run(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.Agent, "agent", "", "Agent CLI to use: claude or codex. Default: ask when both are installed")
	cmd.Flags().StringVar(&opts.Dir, "dir", ".", "Repository to onboard")

	return cmd
}

func run(ctx context.Context, opts *options) error {
	repo, err := filepath.Abs(opts.Dir)
	if err != nil {
		return err
	}
	configPath, err := absoluteConfigPath(opts.ConfigPath)
	if err != nil {
		return err
	}
	signals, err := Scan(repo)
	if err != nil {
		return err
	}

	deployment, ok := opts.Config.GetActiveDeployment()
	if !ok {
		return fmt.Errorf("no active deployment")
	}
	dir, err := sessionDir(time.Now())
	if err != nil {
		return err
	}
	contextPath, err := writeContext(dir, repo, login{
		Alias:          opts.Config.ActiveDeployment,
		OrganizationID: deployment.OrganizationID,
		URL:            deployment.URL,
		ConfigPath:     configPath,
	}, signals)
	if err != nil {
		return err
	}

	selected, err := chooseAgent(opts)
	if err != nil {
		return err
	}
	if selected.command == "" {
		fmt.Fprintf(opts.IO.ErrOut(), heredoc.Doc(`
			No supported agent CLI (claude, codex) was found on this machine.

			The onboarding brief was written to:
			  %s

			Paste that file into any coding agent to onboard this repository by hand.
		`), contextPath)
		return nil
	}

	fmt.Fprintf(opts.IO.ErrOut(), "Starting %s. It will follow the axiom-setup-observability skill.\n\n", selected.label)
	return runAgent(ctx, selected.command, repo, launchPrompt(contextPath))
}

func chooseAgent(opts *options) (agent, error) {
	installed := installedAgents()
	selected, needPrompt, err := resolveAgent(installed, opts.Agent)
	if err != nil || !needPrompt {
		return selected, err
	}
	if !opts.IO.IsStdinTTY() || !opts.IO.IsStdoutTTY() {
		return agent{}, errAgentChoiceRequired
	}

	labels := make([]string, len(installed))
	for i, candidate := range installed {
		labels[i] = candidate.label
	}
	var label string
	if err := survey.AskOne(&survey.Select{
		Message: "Run setup with:",
		Options: labels,
	}, &label, opts.IO.SurveyIO()); err != nil {
		return agent{}, err
	}
	for _, candidate := range installed {
		if candidate.label == label {
			return candidate, nil
		}
	}
	return agent{}, fmt.Errorf("unknown agent %q", label)
}
