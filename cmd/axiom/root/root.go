package root

import (
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/internal/config"
	"github.com/axiomhq/cli/pkg/version"

	// Additional commands
	authCmd "github.com/axiomhq/cli/cmd/axiom/auth"
	completionCmd "github.com/axiomhq/cli/cmd/axiom/completion"
	versionCmd "github.com/axiomhq/cli/cmd/axiom/version"

	// Code command
	datasetCmd "github.com/axiomhq/cli/cmd/axiom/dataset"
	ingestCmd "github.com/axiomhq/cli/cmd/axiom/ingest"
	integrateCmd "github.com/axiomhq/cli/cmd/axiom/integrate"
	queryCmd "github.com/axiomhq/cli/cmd/axiom/query"
	streamCmd "github.com/axiomhq/cli/cmd/axiom/stream"
)

// NewRootCmd creates and returns the root command.
func NewRootCmd(f *cmdutil.Factory) *cobra.Command {
	var (
		configFile       string
		backendOverwrite string
	)

	cmd := &cobra.Command{
		Use:   "axiom <command> <subcommand>",
		Short: "Axiom CLI",
		Long:  `The power of Axiom on the command line.`,

		SilenceErrors: true,
		SilenceUsage:  true,

		// Breaks completion when uncommented.

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
			if backendOverwrite != "" {
				f.Config.ActiveBackend = backendOverwrite
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

	// Active backend overwrite
	cmd.PersistentFlags().StringVarP(&backendOverwrite, "backend", "B", "", "Backend to use by default")

	// Additional commands
	cmd.AddCommand(authCmd.NewAuthCmd(f))
	cmd.AddCommand(completionCmd.NewCompletionCmd())
	cmd.AddCommand(versionCmd.NewVersionCmd(version.Print("Axiom CLI")))

	// Core commands
	cmd.AddCommand(datasetCmd.NewDatasetCmd(f))
	cmd.AddCommand(ingestCmd.NewIngestCmd(f))
	cmd.AddCommand(integrateCmd.NewIntegrateCmd(f))
	cmd.AddCommand(queryCmd.NewQueryCmd(f))
	cmd.AddCommand(streamCmd.NewStreamCmd(f))

	// Help topics
	cmd.AddCommand(newHelpTopic("environment"))

	return cmd
}
