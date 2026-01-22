package auth

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/MakeNowJust/heredoc"
	"github.com/axiomhq/axiom-go/axiom"
	"github.com/spf13/cobra"

	"github.com/axiomhq/cli/internal/client"
	"github.com/axiomhq/cli/internal/cmdutil"
	"github.com/axiomhq/cli/internal/config"
)

func newSwitchOrgCmd(f *cmdutil.Factory) *cobra.Command {
	var orgID string

	cmd := &cobra.Command{
		Use:   "switch-org [<org-id>]",
		Short: "Switch the organization",
		Long: heredoc.Doc(`
			Select the Organization to use for the active deployment and persist
			the choice in the configuration file.
		`),

		DisableFlagsInUseLine: true,

		Args:              cmdutil.PopulateFromArgs(f, &orgID),
		ValidArgsFunction: cmdutil.OrganizationCompletionFunc(f),

		Example: heredoc.Doc(`
			# Select the organization to use by default:
			$ axiom auth switch-org
			
			# Specify the organization to use by default:
			$ axiom auth switch-org my-org-123
		`),

		PreRunE: cmdutil.ChainRunFuncs(
			cmdutil.AsksForSetup(f, NewLoginCmd(f)),
			cmdutil.NeedsDeployments(f),
			cmdutil.NeedsActiveDeployment(f),
			cmdutil.NeedsPersonalAccessToken(f),
		),

		RunE: func(cmd *cobra.Command, _ []string) error {
			if orgID == "" {
				organizations, err := getOrganizations(cmd.Context(), f)
				if err != nil {
					return err
				}

				organizationNames := make([]string, len(organizations))
				for i, organization := range organizations {
					organizationNames[i] = organization.Name
				}

				var organizationName string
				if err := survey.AskOne(&survey.Select{
					Message: "Which organization to use?",
					Default: organizationNames[0],
					Options: organizationNames,
					Description: func(_ string, idx int) string {
						return organizations[idx].ID
					},
				}, &organizationName, f.IO.SurveyIO()); err != nil {
					return err
				}

				for i, organization := range organizations {
					if organization.Name == organizationName {
						orgID = organizations[i].ID
						break
					}
				}
			}

			// A requirement for this command to execute is the presence of an
			// active deployment, so no need to check for existence.
			activeDeployment, _ := f.Config.GetActiveDeployment()

			client, err := client.New(cmd.Context(), activeDeployment.URL, activeDeployment.Token, orgID, "", "", f.Config.Insecure)
			if err != nil {
				return err
			}

			organization, err := client.Organizations.Get(cmd.Context(), orgID)
			if err != nil {
				return err
			}

			f.Config.Deployments[f.Config.ActiveDeployment] = config.Deployment{
				URL:            activeDeployment.URL,
				Token:          activeDeployment.Token,
				OrganizationID: orgID,
			}

			if err := f.Config.Write(); err != nil {
				return err
			}

			cs := f.IO.ColorScheme()
			fmt.Fprintf(f.IO.ErrOut(), "%s Now using organization %s by default\n",
				cs.SuccessIcon(), cs.Bold(organization.Name))

			return nil
		},
	}

	return cmd
}

func getOrganizations(ctx context.Context, f *cmdutil.Factory) ([]*axiom.Organization, error) {
	client, err := f.Client(ctx)
	if err != nil {
		return nil, err
	}

	stop := f.IO.StartActivityIndicator()
	defer stop()

	organizations, err := client.Organizations.List(ctx)
	if err != nil {
		return nil, err
	}

	sort.Slice(organizations, func(i, j int) bool {
		return strings.ToLower(organizations[i].Name) < strings.ToLower(organizations[j].Name)
	})

	stop()

	return organizations, nil
}
