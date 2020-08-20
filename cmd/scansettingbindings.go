package main

import "k8s.io/apimachinery/pkg/runtime/schema"

type ScanSettingBindingHelper struct {
	opts *FCROptions
	gvk  schema.GroupVersionResource
	name string
}

func NewScanSettingBindingHelper(opts *FCROptions, name string) *ScanSettingBindingHelper {
	return &ScanSettingBindingHelper{
		opts: opts,
		name: name,
		gvk: schema.GroupVersionResource{
			Group:    cmpAPIGroup,
			Version:  cmpResourceVersion,
			Resource: "scansettingbindings",
		},
	}
}

func (h *ScanSettingBindingHelper) Handle() error {
	return nil
}
