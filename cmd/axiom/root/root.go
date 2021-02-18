package root

import (
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/internal/config"
	"github.com/axiomhq/cli/pkg/version"

	// Core commands
	ingestCmd "github.com/axiomhq/cli/cmd/axiom/ingest"
	streamCmd "github.com/axiomhq/cli/cmd/axiom/stream"

	// Management commands
	configCmd "github.com/axiomhq/cli/cmd/axiom/config"
	datasetCmd "github.com/axiomhq/cli/cmd/axiom/dataset"

	// Additional commands
	authCmd "github.com/axiomhq/cli/cmd/axiom/auth"
	completionCmd "github.com/axiomhq/cli/cmd/axiom/completion"
	versionCmd "github.com/axiomhq/cli/cmd/axiom/version"
)

// NewRootCmd creates and returns the root command.
func NewRootCmd(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "axiom <command> <subcommand>",
		Short: "Axiom CLI",
		Long:  "The power of Axiom on the command-line.",

		SilenceErrors: true,
		SilenceUsage:  true,

		Example: heredoc.Doc(`
			$ axiom auth login
			$ axiom version
			$ cat /var/log/nginx/*.log | axiom ingest nginx-logs
		`),

		Annotations: map[string]string{
			"help:credentials": heredoc.Doc(`
				See 'axiom help credentials' for help and guidance on authentication.
			`),
			"help:environment": heredoc.Doc(`
				See 'axiom help environment' for the list of supported environment variables.
			`),
		},

		PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
			if fl := cmd.Flag("config"); fl.Changed {
				if f.Config, err = config.Load(fl.Value.String()); err != nil {
					return err
				}
			}

			if fl := cmd.Flag("deployment"); fl.Changed {
				f.Config.ActiveDeployment = fl.Value.String()
			}
			if fl := cmd.Flag("org-id"); fl.Changed {
				f.Config.OrganizationIDOverride = fl.Value.String()
			}
			if fl := cmd.Flag("token"); fl.Changed {
				f.Config.TokenOverride = fl.Value.String()
			}
			if fl := cmd.Flag("url"); fl.Changed {
				f.Config.URLOverride = fl.Value.String()
			}

			f.Config.Insecure = cmd.Flag("insecure").Changed

			return nil
		},
	}

	// IO
	cmd.SetIn(f.IO.In())
	cmd.SetOut(f.IO.Out())
	cmd.SetErr(f.IO.ErrOut())

	// Version
	cmd.Flags().BoolP("version", "v", false, "Show axiom version")
	cmd.SetVersionTemplate("{{ .Short }} version {{ .Version }}\n")
	cmd.Version = version.Release()

	// Help & usage
	cmd.PersistentFlags().BoolP("help", "h", false, "Show help for command")
	cmd.SetHelpFunc(rootHelpFunc(f.IO))
	cmd.SetUsageFunc(rootUsageFunc)
	cmd.SetFlagErrorFunc(rootFlagErrrorFunc)

	// Overrides
	cmd.PersistentFlags().StringP("config", "C", "", "Path to configuration file to use")
	cmd.PersistentFlags().StringP("deployment", "D", "", "Deployment to use")
	cmd.PersistentFlags().StringP("org-id", "O", os.Getenv("AXM_ORG_ID"), "Organization ID to use (only valid for Axiom Cloud)")
	cmd.PersistentFlags().StringP("token", "T", os.Getenv("AXM_TOKEN"), "Token to use")
	cmd.PersistentFlags().StringP("url", "U", os.Getenv("AXM_URL"), "Url to use")
	cmd.PersistentFlags().BoolP("insecure", "I", false, "Bypass certificate validation")

	// Core commands
	cmd.AddCommand(ingestCmd.NewIngestCmd(f))
	cmd.AddCommand(streamCmd.NewStreamCmd(f))

	// Management commands
	cmd.AddCommand(configCmd.NewConfigCmd(f))
	cmd.AddCommand(datasetCmd.NewDatasetCmd(f))

	// Additional commands
	cmd.AddCommand(authCmd.NewAuthCmd(f))
	cmd.AddCommand(completionCmd.NewCompletionCmd(f))
	cmd.AddCommand(versionCmd.NewVersionCmd(f, version.Print("Axiom CLI")))

	// Help topics
	cmd.AddCommand(newHelpTopic(f.IO, "credentials"))
	cmd.AddCommand(newHelpTopic(f.IO, "environment"))

	return cmd
}
