package rerunnow

import (
	"context"
	"fmt"

	"github.com/JAORMX/oc-compliance/internal/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	// FIXME(jaosorior): Normally I would have imported this from the compliance-operator
	// But that seems overkill for one dependency
	rescanAnnotation = "compliance.openshift.io/rescan"
)

type ComplianceScanHelper struct {
	kuser common.KubeClientUser
	gvk   schema.GroupVersionResource
	kind  string
	name  string
}

func NewComplianceScanHelper(kuser common.KubeClientUser, name string) common.ObjectHelper {
	return &ComplianceScanHelper{
		kuser: kuser,
		name:  name,
		kind:  "ComplianceScan",
		gvk: schema.GroupVersionResource{
			Group:    common.CmpAPIGroup,
			Version:  common.CmpResourceVersion,
			Resource: "compliancescans",
		},
	}
}

func (h *ComplianceScanHelper) Handle() error {
	scan, err := h.kuser.DynamicClient().Resource(h.gvk).Namespace(h.kuser.GetNamespace()).Get(context.TODO(), h.name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	annotations := scan.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	annotations[rescanAnnotation] = ""
	scan.SetAnnotations(annotations)

	fmt.Printf("Re-running scan '%s/%s'\n", h.kuser.GetNamespace(), h.name)
	_, err = h.kuser.DynamicClient().Resource(h.gvk).Namespace(h.kuser.GetNamespace()).Update(context.TODO(), scan, metav1.UpdateOptions{})
	return err
}
