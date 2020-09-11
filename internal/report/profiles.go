package report

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/olekukonko/tablewriter"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/JAORMX/oc-compliance/internal/common"
)

const controlAnnotationPrefix = "control.compliance.openshift.io/"

type ProfileHelper struct {
	kuser   common.KubeClientUser
	gvk     schema.GroupVersionResource
	rulegvk schema.GroupVersionResource
	kind    string
	name    string
}

func NewProfileHelper(kuser common.KubeClientUser, name string) common.ObjectHelper {
	return &ProfileHelper{
		kuser: kuser,
		name:  name,
		kind:  "Profile",
		gvk: schema.GroupVersionResource{
			Group:    common.CmpAPIGroup,
			Version:  common.CmpResourceVersion,
			Resource: "profiles",
		},
		rulegvk: schema.GroupVersionResource{
			Group:    common.CmpAPIGroup,
			Version:  common.CmpResourceVersion,
			Resource: "rules",
		},
	}
}

func (h *ProfileHelper) Handle() error {
	results := map[string][]string{}

	p, err := h.kuser.DynamicClient().Resource(h.gvk).Namespace(h.kuser.GetNamespace()).Get(context.TODO(), h.name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	rules, err := h.getRules(p)
	if err != nil {
		return err
	}
	for _, rulename := range rules {
		r, err := h.kuser.DynamicClient().Resource(h.rulegvk).Namespace(h.kuser.GetNamespace()).Get(context.TODO(), rulename, metav1.GetOptions{})
		if err != nil {
			return err
		}
		annotations := r.GetAnnotations()
		if annotations == nil {
			continue
		}

		for key, value := range annotations {
			if strings.HasPrefix(key, controlAnnotationPrefix) {
				benchmark := strings.Split(key, "/")[1]
				_, ok := results[benchmark]
				if !ok {
					results[benchmark] = make([]string, 0)
				}
				insertControlEntries(results, benchmark, value)
			}
		}
	}

	h.render(results)
	return nil
}

func (h *ProfileHelper) getRules(obj *unstructured.Unstructured) ([]string, error) {
	rules, found, err := unstructured.NestedStringSlice(obj.Object, "rules")
	if err != nil {
		return nil, fmt.Errorf("Unable to get rules of %s/%s of type %s: %s", obj.GetNamespace(), obj.GetName(), h.kind, err)
	}
	if !found {
		return nil, fmt.Errorf("%s/%s of type %s: has no rules", obj.GetNamespace(), h.name, h.kind)
	}
	return rules, nil
}

func (h *ProfileHelper) render(res map[string][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Framework", "Controls"})
	table.SetAutoMergeCells(true)
	table.SetRowLine(true)

	for benchmark := range res {
		sort.Strings(res[benchmark])
	}

	for benchmark, controls := range res {
		for _, control := range controls {
			table.Append([]string{benchmark, control})
		}
	}
	table.Render()
}

func insertControlEntries(res map[string][]string, benchmark string, rawcontrols string) {
	controls := strings.Split(rawcontrols, ";")
	newControls := []string{}
	for _, incomingctrl := range controls {
		controlFound := false
		if incomingctrl == "" {
			fmt.Printf("WARNING: empty control in %s", rawcontrols)
		}
		for _, existingctrl := range res[benchmark] {
			// If we already have the control in the list; let's skip it
			if incomingctrl == existingctrl {
				controlFound = true
				break
			}
		}
		if !controlFound {
			newControls = append(newControls, incomingctrl)
		}
	}
	res[benchmark] = append(res[benchmark], newControls...)
}
