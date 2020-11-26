package root

import (
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
	integrateCmd "github.com/axiomhq/cli/cmd/axiom/integrate"

	// Additional commands
	authCmd "github.com/axiomhq/cli/cmd/axiom/auth"
	completionCmd "github.com/axiomhq/cli/cmd/axiom/completion"
	versionCmd "github.com/axiomhq/cli/cmd/axiom/version"
)

// NewRootCmd creates and returns the root command.
func NewRootCmd(f *cmdutil.Factory) *cobra.Command {
	var (
		configFile          string
		deploymentOverwrite string
	)

	cmd := &cobra.Command{
		Use:   "axiom <command> <subcommand>",
		Short: "Axiom CLI",
		Long:  "The power of Axiom on the command-line.",

		SilenceErrors: true,
		SilenceUsage:  true,

		Example: heredoc.Doc(`
			$ axiom auth login
			$ cat /var/log/nginx/*.log | axiom ingest -d nginx-logs
			$ axiom query -d nginx-logs
		`),

		Annotations: map[string]string{
			"help:environment": heredoc.Doc(`
				See 'axiom help environment' for the list of supported environment variables.
			`),
		},

		PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
			if configFile != "" {
				f.Config, err = config.Load(configFile)
			}
			if deploymentOverwrite != "" {
				f.Config.ActiveDeployment = deploymentOverwrite
			}
			return err
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

	// Configuration file overwrite
	cmd.PersistentFlags().StringVarP(&configFile, "config", "C", "", "Path to configuration file to use")

	// Active deployment overwrite
	cmd.PersistentFlags().StringVarP(&deploymentOverwrite, "deployment", "D", "", "Deployment to use by default")

	// Core commands
	cmd.AddCommand(ingestCmd.NewIngestCmd(f))
	cmd.AddCommand(streamCmd.NewStreamCmd(f))

	// Management commands
	cmd.AddCommand(configCmd.NewConfigCmd(f))
	cmd.AddCommand(datasetCmd.NewDatasetCmd(f))
	cmd.AddCommand(integrateCmd.NewIntegrateCmd(f))

	// Additional commands
	cmd.AddCommand(authCmd.NewAuthCmd(f))
	cmd.AddCommand(completionCmd.NewCompletionCmd(f))
	cmd.AddCommand(versionCmd.NewVersionCmd(f, version.Print("Axiom CLI")))

	// Help topics
	cmd.AddCommand(newHelpTopic(f.IO, "environment"))

	return cmd
}
