package main

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ComplianceSuiteHelper struct {
	opts *FCROptions
	gvk  schema.GroupVersionResource
	name string
	kind string
}

func NewComplianceSuiteHelper(opts *FCROptions, name string) *ComplianceSuiteHelper {
	return &ComplianceSuiteHelper{
		opts: opts,
		name: name,
		kind: "ComplianceSuite",
		gvk: schema.GroupVersionResource{
			Group:    cmpAPIGroup,
			Version:  cmpResourceVersion,
			Resource: "compliancesuites",
		},
	}
}

func (h *ComplianceSuiteHelper) Handle() error {
	// Get target resource
	res, err := h.opts.dynclient.Resource(h.gvk).Namespace(h.opts.namespace).Get(context.TODO(), h.name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("Unable to get resource %s/%s of type %s: %s", h.opts.namespace, h.name, h.kind, err)
	}

	// Get needed data
	scanNames, err := h.getScanNames(res)
	if err != nil {
		return err
	}

	fmt.Printf("Fetching results for %s scans: %s\n", h.name, strings.Join(scanNames, ", "))

	for _, scanName := range scanNames {
		scanDir := path.Join(h.opts.outputPath, scanName)
		if err := os.Mkdir(scanDir, 0700); err != nil {
			return fmt.Errorf("Unable to create directory %s: %s", scanDir, err)
		}
		helper := NewComplianceScanHelper(h.opts, scanName, scanDir)
		if err = helper.Handle(); err != nil {
			return fmt.Errorf("Unable to process results from suite %s: %s", h.name, err)
		}
	}
	return nil
}

func (h *ComplianceSuiteHelper) getScanNames(obj *unstructured.Unstructured) ([]string, error) {
	scanNames := []string{}
	scans, found, err := unstructured.NestedSlice(obj.Object, "spec", "scans")
	if err != nil {
		return nil, fmt.Errorf("Unable to get scans of '%s/%s' of type %s", h.opts.namespace, h.name, h.kind)
	}
	if !found {
		return nil, fmt.Errorf("'%s/%s' of type %s has no scans in spec", h.opts.namespace, h.name, h.kind)
	}
	for _, rawscan := range scans {
		scan, ok := rawscan.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("Couldn't parse scan: %v", rawscan)
		}
		name, found, err := unstructured.NestedString(scan, "name")
		if err != nil {
			return nil, fmt.Errorf("Unable to get scan name of '%s/%s' of type %s", h.opts.namespace, h.name, h.kind)
		}
		if !found {
			return nil, fmt.Errorf("'%s/%s' of type %s has no scan name in spec", h.opts.namespace, h.name, h.kind)
		}

		scanNames = append(scanNames, name)
	}
	return scanNames, nil
}
