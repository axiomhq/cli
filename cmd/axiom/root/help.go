// Derived from https://github.com/cli/cli/blob/trunk/pkg/cmd/root/help.go
package root

import (
	"errors"
	"fmt"
	"strings"

	"github.com/muesli/reflow/dedent"
	"github.com/muesli/reflow/indent"
	"github.com/muesli/reflow/padding"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/pkg/terminal"
)

func rootUsageFunc(cmd *cobra.Command) error {
	cmd.Printf("Usage:  %s", cmd.UseLine())

	subCommands := cmd.Commands()
	if len(subCommands) > 0 {
		cmd.Print("\n\nAvailable commands:\n")
		for _, c := range subCommands {
			if c.Hidden {
				continue
			}
			cmd.Printf("  %s\n", c.Name())
		}
		return nil
	}

	flagUsages := cmd.LocalFlags().FlagUsages()
	if flagUsages != "" {
		cmd.Println("\n\nFlags:")
		cmd.Print(indent.String(dedent.String(flagUsages), 2))
	}
	return nil
}

func rootFlagErrrorFunc(cmd *cobra.Command, err error) error {
	if errors.Is(err, pflag.ErrHelp) {
		return err
	}
	return cmdutil.NewFlagError(err)
}

// Display helpful error message in case subcommand name was mistyped.
// This matches Cobra's behavior for root command, which Cobra
// confusingly doesn't apply to nested commands.
func nestedSuggestFunc(cmd *cobra.Command, arg string) {
	cmd.Printf("Error: unknown command %q for %q\n", arg, cmd.CommandPath())

	var candidates []string
	if arg == "help" {
		candidates = []string{"--help"}
	} else {
		if cmd.SuggestionsMinimumDistance <= 0 {
			cmd.SuggestionsMinimumDistance = 2
		}
		candidates = cmd.SuggestionsFor(arg)
	}

	if len(candidates) > 0 {
		cmd.Print("\nDid you mean this?\n")
		for _, c := range candidates {
			cmd.Printf("\t%s\n", c)
		}
	}

	cmd.Print("\n")
	_ = rootUsageFunc(cmd)
}

func isRootCmd(cmd *cobra.Command) bool {
	return cmd != nil && !cmd.HasParent()
}

// HasFailed signals that the help func has failed.
func HasFailed() bool {
	return hasFailed
}

var hasFailed bool

func rootHelpFunc(io *terminal.IO) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		if isRootCmd(cmd.Parent()) && len(args) >= 2 && args[1] != "--help" && args[1] != "-h" {
			nestedSuggestFunc(cmd, args[1])
			hasFailed = true
			return
		}

		coreCommands := []string{}
		managementCommands := []string{}
		additionalCommands := []string{}
		for _, c := range cmd.Commands() {
			if c.Short == "" || c.Hidden {
				continue
			}

			s := padding.String(c.Name()+":", uint(c.NamePadding()+1)) + c.Short
			if _, ok := c.Annotations["IsCore"]; ok {
				coreCommands = append(coreCommands, s)
			} else if _, ok := c.Annotations["IsManagement"]; ok {
				managementCommands = append(managementCommands, s)
			} else {
				additionalCommands = append(additionalCommands, s)
			}
		}

		// If there are no core commands, assume everything is a core command
		if len(coreCommands) == 0 {
			coreCommands = additionalCommands
			additionalCommands = []string{}
		}

		type helpEntry struct {
			Title string
			Body  string
		}

		helpEntries := []helpEntry{}
		if cmd.Long != "" {
			helpEntries = append(helpEntries, helpEntry{"", cmd.Long})
		} else if cmd.Short != "" {
			helpEntries = append(helpEntries, helpEntry{"", cmd.Short})
		}
		helpEntries = append(helpEntries, helpEntry{"USAGE", cmd.UseLine()})
		if len(cmd.Aliases) > 0 {
			helpEntries = append(helpEntries, helpEntry{"ALIASES", strings.Join(cmd.Aliases, "\n")})
		}
		if len(coreCommands) > 0 {
			helpEntries = append(helpEntries, helpEntry{"CORE COMMANDS", strings.Join(coreCommands, "\n")})
		}
		if len(managementCommands) > 0 {
			helpEntries = append(helpEntries, helpEntry{"MANAGEMENT COMMANDS", strings.Join(managementCommands, "\n")})
		}
		if len(additionalCommands) > 0 {
			helpEntries = append(helpEntries, helpEntry{"ADDITIONAL COMMANDS", strings.Join(additionalCommands, "\n")})
		}

		flagUsages := cmd.LocalFlags().FlagUsages()
		if flagUsages != "" {
			helpEntries = append(helpEntries, helpEntry{"FLAGS", dedent.String(flagUsages)})
		}
		inheritedFlagUsages := cmd.InheritedFlags().FlagUsages()
		if inheritedFlagUsages != "" {
			helpEntries = append(helpEntries, helpEntry{"INHERITED FLAGS", dedent.String(inheritedFlagUsages)})
		}
		if _, ok := cmd.Annotations["help:arguments"]; ok {
			helpEntries = append(helpEntries, helpEntry{"ARGUMENTS", cmd.Annotations["help:arguments"]})
		}
		if cmd.Example != "" {
			helpEntries = append(helpEntries, helpEntry{"EXAMPLES", cmd.Example})
		}
		if _, ok := cmd.Annotations["help:environment"]; ok {
			helpEntries = append(helpEntries, helpEntry{"ENVIRONMENT VARIABLES", cmd.Annotations["help:environment"]})
		}
		helpEntries = append(helpEntries, helpEntry{"LEARN MORE", `
Use 'axiom <command> <subcommand> --help' for more information about a command.
Read the manual at https://docs.axiom.co/cli`})
		if _, ok := cmd.Annotations["help:feedback"]; ok {
			helpEntries = append(helpEntries, helpEntry{"FEEDBACK", cmd.Annotations["help:feedback"]})
		}

		for _, e := range helpEntries {
			if e.Title != "" {
				// If there is a title, add indentation to each line in the body.
				fmt.Fprintln(io.Out(), io.ColorScheme().Bold(e.Title))
				fmt.Fprintln(io.Out(), indent.String(strings.Trim(e.Body, "\r\n"), 2))
			} else {
				// If there is no title print the body as is.
				fmt.Fprintln(io.Out(), e.Body)
			}
			fmt.Fprintln(io.Out())
		}
	}
}
