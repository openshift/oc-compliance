package fetchraw

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/JAORMX/oc-compliance/internal/common"
)

type FetchRawOptions struct {
	ConfigFlags *genericclioptions.ConfigFlags

	kuser common.KubeClientUser

	OutputPath string

	args []string

	helper common.ObjectHelper

	genericclioptions.IOStreams
}

// Complete sets all information required for updating the current context
func (o *FetchRawOptions) Complete(cmd *cobra.Command, args []string) error {
	o.args = args

	// Takes precedence
	givenNamespace, err := cmd.Flags().GetString("namespace")
	if err != nil {
		return err
	}
	o.kuser, err = common.NewKubeClientUser(o.ConfigFlags, givenNamespace)
	if err != nil {
		return err
	}
	return nil
}

// Validate ensures that all required arguments and flag values are provided
func (o *FetchRawOptions) Validate() error {
	finfo, err := os.Stat(o.OutputPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("The directory at path '%s' doesn't exist", o.OutputPath)
	}

	if !finfo.IsDir() {
		return fmt.Errorf("The output path must be a directory")
	}

	objtype, objname, err := common.ValidateObjectArgs(o.args)
	if err != nil {
		return err
	}

	switch objtype {
	case common.ScanSettingBinding:
		o.helper = NewScanSettingBindingHelper(o, o.kuser, objname, o.OutputPath)
	case common.ComplianceSuite:
		o.helper = NewComplianceSuiteHelper(o, o.kuser, objname, o.OutputPath)
	case common.ComplianceScan:
		o.helper = NewComplianceScanHelper(o, o.kuser, objname, o.OutputPath)
	}

	return nil
}

func (o *FetchRawOptions) Run() error {
	return o.helper.Handle()
}
