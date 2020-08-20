package main

import "k8s.io/apimachinery/pkg/runtime/schema"

type ComplianceSuiteHelper struct {
	opts *FCROptions
	gvk  schema.GroupVersionResource
	name string
}

func NewComplianceSuiteHelper(opts *FCROptions, name string) *ComplianceSuiteHelper {
	return &ComplianceSuiteHelper{
		opts: opts,
		name: name,
		gvk: schema.GroupVersionResource{
			Group:    cmpAPIGroup,
			Version:  cmpResourceVersion,
			Resource: "compliancesuites",
		},
	}
}

func (h *ComplianceSuiteHelper) Handle() error {
	return nil
}
