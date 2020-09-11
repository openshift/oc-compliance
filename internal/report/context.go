package report

import (
	"fmt"

	"github.com/JAORMX/oc-compliance/internal/common"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type ReportContext struct {
	common.CommandContext
}

func NewReportContext(streams genericclioptions.IOStreams) *ReportContext {
	return &ReportContext{
		CommandContext: common.CommandContext{
			ConfigFlags: genericclioptions.NewConfigFlags(true),
			IOStreams:   streams,
		},
	}
}

// Validate ensures that all required arguments and flag values are provided
func (o *ReportContext) Validate() error {
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

func (o *ReportContext) Run() error {
	return o.Helper.Handle()
}
