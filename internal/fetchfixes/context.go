package fetchfixes

import (
	"fmt"
	"os"

	"github.com/JAORMX/oc-compliance/internal/common"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type FetchFixesContext struct {
	common.CommandContext

	OutputPath string
}

func NewFetchFixesContext(streams genericclioptions.IOStreams) *FetchFixesContext {
	return &FetchFixesContext{
		CommandContext: common.CommandContext{
			ConfigFlags: genericclioptions.NewConfigFlags(true),
			IOStreams:   streams,
		},
	}
}

// Validate ensures that all required arguments and flag values are provided
func (o *FetchFixesContext) Validate() error {
	finfo, err := os.Stat(o.OutputPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("The directory at path '%s' doesn't exist", o.OutputPath)
	}

	if !finfo.IsDir() {
		return fmt.Errorf("The output path must be a directory")
	}

	objtype, objname, err := common.ValidateObjectArgs(o.Args)
	if err != nil {
		return err
	}

	switch objtype {
	case common.Rule:
		o.Helper = NewRuleHelper(o.Kuser, objname, o.OutputPath, o.IOStreams)
	case common.Profile:
		o.Helper = NewProfileHelper(o.Kuser, objname, o.OutputPath, o.IOStreams)
	default:
		return fmt.Errorf("Invalid object type for this command")
	}
	return nil
}

func (o *FetchFixesContext) Run() error {
	return o.Helper.Handle()
}
