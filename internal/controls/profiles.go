package controls

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/olekukonko/tablewriter"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/openshift/oc-compliance/internal/common"
)

const controlAnnotationPrefix = "control.compliance.openshift.io/"

type ProfileHelper struct {
	kuser   common.KubeClientUser
	gvk     schema.GroupVersionResource
	rulegvk schema.GroupVersionResource
	kind    string
	name    string
	genericclioptions.IOStreams
}

func NewProfileHelper(kuser common.KubeClientUser, name string, streams genericclioptions.IOStreams) common.ObjectHelper {
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
		IOStreams: streams,
	}
}

func (h *ProfileHelper) Handle() error {
	results := map[string][]string{}

	p, err := h.kuser.DynamicClient().Resource(h.gvk).Namespace(h.kuser.GetNamespace()).Get(context.TODO(), h.name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	rules, err := common.GetRulesFromProfile(p)
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

func (h *ProfileHelper) render(res map[string][]string) {
	table := tablewriter.NewWriter(h.Out)
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
