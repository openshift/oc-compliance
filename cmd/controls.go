package main

import (
	"fmt"
	"os"

	"github.com/JAORMX/oc-compliance/internal/controls"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func init() {
	controlsCmd := NewCmdControls(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	rootCmd.AddCommand(controlsCmd)
}

func NewCmdControls(streams genericclioptions.IOStreams) *cobra.Command {
	var (
		controlsExamples = `
  # Get a controls for the profile
  %[1]s %[2]s profile [resource name]
`
	)

	ctx := controls.NewControlsContext(streams)
	cmd := &cobra.Command{
		Use:          "controls [object] [object name]",
		Short:        "Get a report of what controls you're complying with",
		Long:         "Get a report of what controls you're complying with",
		Example:      fmt.Sprintf(controlsExamples, "oc compliance", "controls"),
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
