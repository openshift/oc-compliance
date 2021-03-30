package fetchfixes

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8sserial "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/openshift/oc-compliance/internal/common"
)

type ComplianceRemediationHelper struct {
	kuser      common.KubeClientUser
	gvk        schema.GroupVersionResource
	kind       string
	name       string
	outputPath string
	genericclioptions.IOStreams
}

func NewComplianceRemediationHelper(kuser common.KubeClientUser, name string, outputPath string, streams genericclioptions.IOStreams) common.ObjectHelper {
	return &ComplianceRemediationHelper{
		kuser: kuser,
		name:  name,
		kind:  "ComplianceRemediation",
		gvk: schema.GroupVersionResource{
			Group:    common.CmpAPIGroup,
			Version:  common.CmpResourceVersion,
			Resource: "complianceremediations",
		},
		outputPath: outputPath,
		IOStreams:  streams,
	}
}

func (h *ComplianceRemediationHelper) Handle() error {
	r, err := h.kuser.DynamicClient().Resource(h.gvk).Namespace(h.kuser.GetNamespace()).Get(context.TODO(), h.name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	current, err := h.getCurrentObject(r)
	if err != nil {
		return err
	}

	yamlSerializer := k8sserial.NewYAMLSerializer(k8sserial.DefaultMetaFactory, nil, nil)

	path, err := common.PersistObjectToYamlFile(r.GetName(), current, h.outputPath, yamlSerializer)
	if err != nil {
		return err
	}

	fmt.Fprintf(h.Out, "Persisted compliance remediation fix to %s\n", path)

	return nil
}

func (h *ComplianceRemediationHelper) getCurrentObject(obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	rem, found, err := unstructured.NestedMap(obj.Object, "spec", "current", "object")
	if err != nil {
		return nil, fmt.Errorf("Unable to get remediations of %s/%s of type %s: %s", obj.GetNamespace(), obj.GetName(), h.kind, err)
	}
	if !found {
		return nil, fmt.Errorf("no found remediations for %s/%s of type %s: %s", obj.GetNamespace(), obj.GetName(), h.kind, err)
	}

	remobj := &unstructured.Unstructured{Object: rem}
	return remobj, nil
}
