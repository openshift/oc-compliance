package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/JAORMX/oc-compliance/internal/viewresult"
)

func init() {
	viewresultCmd := NewCmdViewResult(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	rootCmd.AddCommand(viewresultCmd)
}

func NewCmdViewResult(streams genericclioptions.IOStreams) *cobra.Command {
	var (
		rerunExamples = `
  # View a result
  %[1]s %[2]s [resource name]
`
	)

	ctx := viewresult.NewViewResultContext(streams)
	cmd := &cobra.Command{
		Use:          "view-result [object] [object name]",
		Short:        "View a result",
		Long:         `'view-result' exposes more information about a compliance result.`,
		Example:      fmt.Sprintf(rerunExamples, "oc compliance", "view-result"),
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
