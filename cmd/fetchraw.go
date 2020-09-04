package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	fetchraw "github.com/JAORMX/oc-compliance/internal/fetchraw"
)

var (
	usageExamples = `
  # Fetch from compliancescan
  %[1]s %[2]s compliancescan [resource name] -o [directory]
  
  # Fetch from compliancesuite
  %[1]s %[2]s compliancesuite [resource name] -o [directory]
  
  # Fetch from scansettingbindings
  %[1]s %[2]s scansettingbindings [resource name] -o [directory]
`

	errNoContext = fmt.Errorf("no context is currently set, use %q to select a new one", "kubectl config use-context <context>")
)

func init() {
	fetchRawCmd := NewCmdFetchRaw(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	rootCmd.AddCommand(fetchRawCmd)
}

func NewCmdFetchRaw(streams genericclioptions.IOStreams) *cobra.Command {
	o := fetchraw.NewFetchRawOptions(streams)

	cmd := &cobra.Command{
		Use:   "fetch-raw [object] [object name] -o [output path]",
		Short: "Download raw compliance results",
		Long: `'fetch-raw' fetches the raw results for a scan or set of scans.

This command allows you to download the raw results from a
compliance scan to a specified directory.`,
		Example:      fmt.Sprintf(usageExamples, "oc compliance", "fetch-raw"),
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

	cmd.Flags().StringVarP(&o.OutputPath, "output", "o", ".", "The path where you want to persist the raw results to")
	o.ConfigFlags.AddFlags(cmd.Flags())

	return cmd
}
