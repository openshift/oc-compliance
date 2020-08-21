package main

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ScanSettingBindingHelper struct {
	opts       *FCROptions
	gvk        schema.GroupVersionResource
	name       string
	kind       string
	outputPath string
}

func NewScanSettingBindingHelper(opts *FCROptions, name, outputPath string) *ScanSettingBindingHelper {
	return &ScanSettingBindingHelper{
		opts:       opts,
		name:       name,
		kind:       "ScanSettingBinding",
		outputPath: outputPath,
		gvk: schema.GroupVersionResource{
			Group:    cmpAPIGroup,
			Version:  cmpResourceVersion,
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
