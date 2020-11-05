package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/cmdutil"
)

func newEditCmd(f *cmdutil.Factory) *cobra.Command {
	var content string

	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit the configuration file",
		Long:  `Open the editor to edit the configuration file.`,

		DisableFlagsInUseLine: true,

		Example: heredoc.Doc(`
			# Open the default configuration file in the configured editor:
			$ axiom config edit
			
			# Open the specified configuration file in the configured editor:
			$ axiom config edit -C /etc/axiom/cli.toml
		`),

		PreRunE: func(*cobra.Command, []string) error {
			f, err := os.Open(f.Config.ConfigFilePath)
			if err != nil {
				return err
			}
			defer f.Close()

			b, err := ioutil.ReadAll(f)
			if err != nil {
				return err
			}
			content = string(b)

			return f.Close()
		},

		RunE: func(*cobra.Command, []string) error {
			if !f.IO.IsStdinTTY() {
				return errors.New("cannot run this command non interactively")
			}

			return survey.AskOne(&survey.Editor{
				Message:       fmt.Sprintf("Edit %s", f.Config.ConfigFilePath),
				FileName:      "*.toml",
				Default:       content,
				HideDefault:   true,
				AppendDefault: true,
			}, &content, f.IO.SurveyIO())
		},

		PostRunE: func(*cobra.Command, []string) error {
			f, err := os.Create(f.Config.ConfigFilePath)
			if err != nil {
				return err
			}
			defer f.Close()

			if _, err = f.WriteString(content); err != nil {
				return err
			}

			return f.Sync()
		},
	}

	return cmd
}
