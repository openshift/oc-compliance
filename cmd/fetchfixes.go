package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	fetchfixes "github.com/JAORMX/oc-compliance/internal/fetchfixes"
)

func init() {
	fetchFixesCmd := NewCmdFetchFixes(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	rootCmd.AddCommand(fetchFixesCmd)
}

func NewCmdFetchFixes(streams genericclioptions.IOStreams) *cobra.Command {
	var (
		usageExamples = `
  # Fetch from a rule
  %[1]s %[2]s rule [resource name] -o [directory]
  
  # Fetch from a profile
  %[1]s %[2]s profile [resource name] -o [directory]
  
  # Fetch from a complianceRemediation
  %[1]s %[2]s complianceremediation [resource name] -o [directory]
`
	)

	o := fetchfixes.NewFetchFixesContext(streams)

	cmd := &cobra.Command{
		Use:   "fetch-fixes [object] [object name] -o [output path]",
		Short: "Download the fixes or remediations",
		Long: `'fetch-fixes' fetches the fixes or remediations from a rule, profile, scan or remediation object.

This command allows you to download the proposed fixes from a
compliance scan or a profile to a specified directory.`,
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
