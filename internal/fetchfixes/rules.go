package fetchfixes

import (
	"context"
	"fmt"
	"os"
	"path"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	k8sserial "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/JAORMX/oc-compliance/internal/common"
)

type RuleHelper struct {
	kuser      common.KubeClientUser
	gvk        schema.GroupVersionResource
	kind       string
	name       string
	outputPath string
	genericclioptions.IOStreams
}

func NewRuleHelper(kuser common.KubeClientUser, name string, outputPath string, streams genericclioptions.IOStreams) common.ObjectHelper {
	return &RuleHelper{
		kuser: kuser,
		name:  name,
		kind:  "Profile",
		gvk: schema.GroupVersionResource{
			Group:    common.CmpAPIGroup,
			Version:  common.CmpResourceVersion,
			Resource: "rules",
		},
		outputPath: outputPath,
		IOStreams:  streams,
	}
}

func (h *RuleHelper) Handle() error {
	r, err := h.kuser.DynamicClient().Resource(h.gvk).Namespace(h.kuser.GetNamespace()).Get(context.TODO(), h.name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	fixes, err := h.getAvailableFixes(r)
	if err != nil {
		return err
	}

	if len(fixes) == 0 {
		fmt.Fprintf(h.Out, "No fixes to persist for rule '%s'\n", r.GetName())
		return nil
	}

	yamlSerializer := k8sserial.NewYAMLSerializer(k8sserial.DefaultMetaFactory, nil, nil)

	// Serialize the objects to yaml
	path := path.Join(h.outputPath, h.name+".yaml")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

	writer := json.YAMLFramer.NewFrameWriter(f)
	needsSuffix := len(fixes) > 1
	for idx, fix := range fixes {
		// Needed for MachineConfigs
		if fix.GetName() == "" {
			setFixName(fix, r.GetName(), idx, needsSuffix)
		}
		if err := yamlSerializer.Encode(fix, writer); err != nil {
			return fmt.Errorf("Couldn't serialize fix from rule '%s': %s", r.GetName(), err)
		}
	}

	if err = f.Sync(); err != nil {
		return err
	}

	fmt.Fprintf(h.Out, "Persisted rule fix to %s\n", path)

	return nil
}

func (h *RuleHelper) getAvailableFixes(obj *unstructured.Unstructured) ([]*unstructured.Unstructured, error) {
	fixes, found, err := unstructured.NestedSlice(obj.Object, "availableFixes")
	if err != nil {
		return nil, fmt.Errorf("Unable to get fixes of %s/%s of type %s: %s", obj.GetNamespace(), obj.GetName(), h.kind, err)
	}
	if !found {
		// This is not a fatal error. Some rules might not have fixes assigned to them
		return nil, nil
	}

	output := []*unstructured.Unstructured{}
	for _, fixIf := range fixes {
		fix, ok := fixIf.(map[string]interface{})
		if !ok {
			fmt.Fprintf(h.ErrOut, "WARNING: Rule '%s/%s' has a malformed fix. Expected a map.", obj.GetNamespace(), obj.GetName())
			continue
		}
		fixobjraw, found, err := unstructured.NestedMap(fix, "fixObject")
		if err != nil {
			fmt.Fprintf(h.ErrOut, "WARNING: Rule '%s/%s' has a malformed fix. Couldn't get 'fixObject' key.", obj.GetNamespace(), obj.GetName())
			continue
		}
		if !found {
			fmt.Fprintf(h.ErrOut, "WARNING: Rule '%s/%s' has a malformed fix. 'fixObject' key not found.", obj.GetNamespace(), obj.GetName())
			continue
		}
		fixobj := &unstructured.Unstructured{Object: fixobjraw}
		output = append(output, fixobj)
	}
	return output, nil
}

func setFixName(obj *unstructured.Unstructured, name string, id int, needsSuffix bool) {
	if !needsSuffix {
		obj.SetName(name)
	} else {
		obj.SetName(fmt.Sprintf("%s-%d", name, id))
	}
}
