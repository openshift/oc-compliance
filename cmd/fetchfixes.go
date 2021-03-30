package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	fetchfixes "github.com/openshift/oc-compliance/internal/fetchfixes"
)

func init() {
	fetchFixesCmd := NewCmdFetchFixes(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	rootCmd.AddCommand(fetchFixesCmd)
}

func NewCmdFetchFixes(streams genericclioptions.IOStreams) *cobra.Command {
	var (
		usageExamples = `
  # Fetch from a rule named "ocp4-api-server-encryption-provider-cipher" into /tmp
  %[1]s %[2]s rule cp4-api-server-encryption-provider-cipher -o /tmp

  # Fetch from a profile named "ocp4-cis" into /tmp
  %[1]s %[2]s profile ocp4-cis -o /tmp

  # Fetch from a complianceRemediation named ocp4-cis-api-server-encryption-provider-cipher into /tmp
  %[1]s %[2]s complianceremediation ocp4-cis-api-server-encryption-provider-cipher -o /tmp
`
	)

	o := fetchfixes.NewFetchFixesContext(streams)

	cmd := &cobra.Command{
		Use:   "fetch-fixes {rule | profile | complianceremediation } <resource-name> -o <output path>",
		Short: "Download the fixes/remediations",
		Long: `'fetch-fixes' fetches the fixes/remediations from a Rule, Profile, or ComplianceRemediation.

This command allows you to download the proposed fixes from a Rule, Profile, or
ComplianceRemediation into a specified directory.`,
		Example:      fmt.Sprintf(usageExamples, "oc compliance", "fetch-fixes"),
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(c, args); err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
				return err
			}
			if err := o.Run(); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&o.OutputPath, "output", "o", ".", "The path where you want to persist the fix objects to")
	o.ConfigFlags.AddFlags(cmd.Flags())

	return cmd
}
