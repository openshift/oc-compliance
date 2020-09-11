package main

import (
	"fmt"
	"os"

	"github.com/JAORMX/oc-compliance/internal/report"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func init() {
	reportCmd := NewCmdReport(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	rootCmd.AddCommand(reportCmd)
}

func NewCmdReport(streams genericclioptions.IOStreams) *cobra.Command {
	var (
		reportExamples = `
  # Get a report for the profile
  %[1]s %[2]s profile [resource name]
`
	)

	ctx := report.NewReportContext(streams)
	cmd := &cobra.Command{
		Use:          "report [object] [object name]",
		Short:        "Get a report of what you're complying with",
		Long:         "Get a report of what you're complying with",
		Example:      fmt.Sprintf(reportExamples, "oc compliance", "report"),
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := ctx.Complete(c, args); err != nil {
				return err
			}
			if err := ctx.Validate(); err != nil {
				return err
			}
			if err := ctx.Run(); err != nil {
				return err
			}

			return nil
		},
	}

	ctx.ConfigFlags.AddFlags(cmd.Flags())
	return cmd
}
