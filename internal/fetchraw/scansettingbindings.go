package fetchraw

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/JAORMX/oc-compliance/internal/common"
)

type ScanSettingBindingHelper struct {
	opts       *FetchRawOptions
	gvk        schema.GroupVersionResource
	name       string
	kind       string
	outputPath string
}

func NewScanSettingBindingHelper(opts *FetchRawOptions, name, outputPath string) *ScanSettingBindingHelper {
	return &ScanSettingBindingHelper{
		opts:       opts,
		name:       name,
		kind:       "ScanSettingBinding",
		outputPath: outputPath,
		gvk: schema.GroupVersionResource{
			Group:    common.CmpAPIGroup,
			Version:  common.CmpResourceVersion,
			Resource: "scansettingbindings",
		},
	}
}

func (h *ScanSettingBindingHelper) Handle() error {
	// Get target resource
	res, err := h.opts.dynclient.Resource(h.gvk).Namespace(h.opts.namespace).Get(context.TODO(), h.name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("Unable to get resource %s/%s of type %s: %s", h.opts.namespace, h.name, h.kind, err)
	}
	suiteName := res.GetName()

	helper := NewComplianceSuiteHelper(h.opts, suiteName, h.outputPath)
	return helper.Handle()
}
