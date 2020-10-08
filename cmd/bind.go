package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	bind "github.com/JAORMX/oc-compliance/internal/bind"
)

func init() {
	bindCmd := NewCmdBind(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	rootCmd.AddCommand(bindCmd)
}

func NewCmdBind(streams genericclioptions.IOStreams) *cobra.Command {
	var (
		usageExamples = `
  # Create a scansettingbinding
  %[1]s %[2]s compliancescan -N [resource name] -s [settings] [objtype/objname]
  
  # Display a scansettingbinding
  %[1]s %[2]s compliancescan --dry-run -N [resource name] -s [settings] [objtype/objname]
`
	)

	o := bind.NewBindContext(streams)

	cmd := &cobra.Command{
		Use:   "bind -n [binding name] -s [settings name] [object] ...",
		Short: "Creates a ScanSettingBinding for the given parameters",
		Long: `'bind' will take the given parameters and create a ScanSettingBinding object.
		
These objects will take the given scanSettings and bind them to the given profiles and tailored profiles.`,
		Example:      fmt.Sprintf(usageExamples, "oc compliance", "bind"),
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

	cmd.Flags().StringVarP(&o.Settings, "settings", "S", "default", "The scan settings to bind the Profiles/TailoredProfiles to")
	cmd.Flags().StringVarP(&o.Name, "name", "N", "", "The name of the binding to create")
	cmd.Flags().BoolVar(&o.DryRun, "dry-run", false, "Output the scansettingbinding that would be created")
	o.ConfigFlags.AddFlags(cmd.Flags())

	return cmd
}
