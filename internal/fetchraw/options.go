package fetchraw

import (
	"fmt"
	"os"
	"strings"

	"github.com/JAORMX/oc-compliance/internal/common"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type FetchRawOptions struct {
	ConfigFlags *genericclioptions.ConfigFlags

	kuser common.KubeClientUser

	OutputPath string

	args []string

	helper ObjectHelper

	genericclioptions.IOStreams
}

type ObjectHelper interface {
	Handle() error
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

	if len(o.args) < 1 {
		return fmt.Errorf("You need to specify at least one object")
	}

	if len(o.args) > 2 {
		return fmt.Errorf("unkown argument(s): %s", o.args[2:])
	}

	var rawobjtype, objname string
	if len(o.args) == 1 {
		objref := strings.Split(o.args[0], "/")
		if len(objref) == 1 {
			return fmt.Errorf("Missing object name")
		}

		if len(objref) > 2 {
			return fmt.Errorf("Malformed reference to object: %s", o.args[0])
		}

		rawobjtype = objref[0]
		objname = objref[1]
	} else {
		rawobjtype = o.args[0]
		objname = o.args[1]
	}

	switch rawobjtype {
	case "ScanSettingBindings", "ScanSettingBinding", "scansettingbindings", "scansettingbinding":
		o.helper = NewScanSettingBindingHelper(o, o.kuser, objname, o.OutputPath)
	case "ComplianceSuites", "ComplianceSuite", "compliancesuites", "compliancesuite":
		o.helper = NewComplianceSuiteHelper(o, o.kuser, objname, o.OutputPath)
	case "ComplianceScans", "ComplianceScan", "compliancescans", "compliancescan":
		o.helper = NewComplianceScanHelper(o, o.kuser, objname, o.OutputPath)
	default:
		return fmt.Errorf("Unkown object type: %s", rawobjtype)
	}

	return nil
}

func (o *FetchRawOptions) Run() error {
	return o.helper.Handle()
}
