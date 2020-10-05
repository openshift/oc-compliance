package controls

import (
	"fmt"

	"github.com/JAORMX/oc-compliance/internal/common"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type ControlsContext struct {
	common.CommandContext
}

func NewControlsContext(streams genericclioptions.IOStreams) *ControlsContext {
	return &ControlsContext{
		CommandContext: common.CommandContext{
			ConfigFlags: genericclioptions.NewConfigFlags(true),
			IOStreams:   streams,
		},
	}
}

// Validate ensures that all required arguments and flag values are provided
func (o *ControlsContext) Validate() error {
	objtype, objname, err := common.ValidateObjectArgs(o.Args)
	if err != nil {
		return err
	}

	switch objtype {
	case common.Profile:
		o.Helper = NewProfileHelper(o.Kuser, objname)
	default:
		return fmt.Errorf("Invalid object type for this command")
	}
	return nil
}

func (o *ControlsContext) Run() error {
	return o.Helper.Handle()
}
