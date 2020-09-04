package main

import (
	"fmt"
	"os"

	"github.com/JAORMX/oc-compliance/internal/rerunnow"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var (
	rerunExamples = `
  # rerun a compliancescan
  %[1]s %[2]s compliancescan [resource name]
  
  # rerun a compliancesuite
  %[1]s %[2]s compliancesuite [resource name]
  
  # rerun a scansettingbindings
  %[1]s %[2]s scansettingbindings [resource name]
`
)

func init() {
	rerunNowCmd := NewCmdRerunNow(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	rootCmd.AddCommand(rerunNowCmd)
}

func NewCmdRerunNow(streams genericclioptions.IOStreams) *cobra.Command {
	ctx := rerunnow.NewReRunNowContext(streams)
	cmd := &cobra.Command{
		Use:          "rerun-now [object] [object name]",
		Short:        "Force a rerun for a scan or set of scans",
		Long:         `'rerun-now' forces a scan or set of scans to be retriggered.`,
		Example:      fmt.Sprintf(rerunExamples, "oc compliance", "rerun-now"),
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
