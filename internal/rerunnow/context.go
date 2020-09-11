package rerunnow

import (
	"fmt"

	"github.com/JAORMX/oc-compliance/internal/common"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type RerunNowContext struct {
	common.CommandContext
}

func NewReRunNowContext(streams genericclioptions.IOStreams) *RerunNowContext {
	return &RerunNowContext{
		CommandContext: common.CommandContext{
			ConfigFlags: genericclioptions.NewConfigFlags(true),
			IOStreams:   streams,
		},
	}
}

// Complete sets all information required for updating the current context
func (o *RerunNowContext) Complete(cmd *cobra.Command, args []string) error {
	o.Args = args

	// Takes precedence
	givenNamespace, err := cmd.Flags().GetString("namespace")
	if err != nil {
		return err
	}
	o.Kuser, err = common.NewKubeClientUser(o.ConfigFlags, givenNamespace)
	if err != nil {
		return err
	}
	return nil
}

// Validate ensures that all required arguments and flag values are provided
func (o *RerunNowContext) Validate() error {
	objtype, objname, err := common.ValidateObjectArgs(o.Args)
	if err != nil {
		return err
	}

	switch objtype {
	case common.ScanSettingBinding:
		o.Helper = NewScanSettingBindingHelper(o.Kuser, objname)
	case common.ComplianceSuite:
		o.Helper = NewComplianceSuiteHelper(o.Kuser, objname)
	case common.ComplianceScan:
		o.Helper = NewComplianceScanHelper(o.Kuser, objname)
	default:
		return fmt.Errorf("Invalid object type for this command")
	}
	return nil
}

func (o *RerunNowContext) Run() error {
	return o.Helper.Handle()
}
