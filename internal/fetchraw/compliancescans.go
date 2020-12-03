package fetchraw

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/cmd/cp"

	"github.com/JAORMX/oc-compliance/internal/common"
)

const (
	rawResultsMountPath = "raw-results"
	cmdLabelKey         = "fetch-compliance-results"
	objNameLabelKey     = "fetch-compliance-results/obj-name"
)

type ComplianceScanHelper struct {
	kuser      common.KubeClientUser
	gvk        schema.GroupVersionResource
	podgvk     schema.GroupVersionResource
	kind       string
	name       string
	outputPath string
	genericclioptions.IOStreams
}

func NewComplianceScanHelper(kuser common.KubeClientUser, name, outputPath string, streams genericclioptions.IOStreams) common.ObjectHelper {
	return &ComplianceScanHelper{
		kuser:      kuser,
		name:       name,
		kind:       "ComplianceScan",
		outputPath: outputPath,
		gvk: schema.GroupVersionResource{
			Group:    common.CmpAPIGroup,
			Version:  common.CmpResourceVersion,
			Resource: "compliancescans",
		},
		podgvk: schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "pods",
		},
		IOStreams: streams,
	}
}

func (h *ComplianceScanHelper) Handle() error {
	// Get target resource
	res, err := h.kuser.DynamicClient().Resource(h.gvk).Namespace(h.kuser.GetNamespace()).Get(context.TODO(), h.name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("Unable to get resource %s/%s of type %s: %s", h.kuser.GetNamespace(), h.name, h.kind, err)
	}

	// Get needed data
	phase, err := h.getScanPhase(res)
	if err != nil {
		return err
	}

	ci, err := h.getCurrentIndex(res)
	if err != nil {
		return err
	}

	if ci == 0 && strings.ToLower(phase) != "done" {
		return fmt.Errorf("No results available yet. Please wait for the scan to be done. Current phase: %s", phase)
	}

	rsRef, err := h.getResultsStorageRef(res)
	if err != nil {
		return err
	}

	claimName, found := rsRef["name"]
	if !found {
		return fmt.Errorf("Malformed raw result storage reference. No name available. Check the %s object's status", h.kind)
	}

	rsnamespace, found := rsRef["namespace"]
	if !found {
		return fmt.Errorf("Malformed raw result storage reference. No namespace available. Check the %s object's status", h.kind)
	}

	// Create extractor pod
	extractorPod := getPVCExtractorPod(res.GetName(), rsnamespace, claimName)
	extractorPod, err = h.kuser.Clientset().CoreV1().Pods(rsnamespace).Create(context.TODO(), extractorPod, metav1.CreateOptions{})
	if err != nil && !kerrors.IsAlreadyExists(err) {
		return err
	}

	// wait for extractor pod
	err = h.waitForExtractorPod(rsnamespace, res.GetName())
	if err != nil {
		return err
	}

	cpopts := cp.NewCopyOptions(h.IOStreams)
	cpopts.Namespace = rsnamespace
	cpopts.ClientConfig = h.kuser.GetConfig()
	cpopts.Clientset = h.kuser.Clientset()

	podName := extractorPod.GetName()
	path := fmt.Sprintf("%s/%d", rawResultsMountPath, ci)
	cpargs := []string{
		fmt.Sprintf("%s/%s:%s", rsnamespace, podName, path),
		h.outputPath,
	}

	// run kubectl cp
	if err = cpopts.Run(cpargs); err != nil {
		return err
	}

	fmt.Fprintf(h.Out, "The raw compliance results are avaliable in the following directory: %s\n", h.outputPath)

	// delete extractor pod
	var zeroGP int64 = 0
	return h.kuser.Clientset().CoreV1().Pods(rsnamespace).Delete(context.TODO(), extractorPod.GetName(), metav1.DeleteOptions{
		GracePeriodSeconds: &zeroGP,
	})
}

func (h *ComplianceScanHelper) getScanPhase(obj *unstructured.Unstructured) (string, error) {
	phase, found, err := unstructured.NestedString(obj.Object, "status", "phase")
	if err != nil {
		return "", fmt.Errorf("Unable to get phase of %s/%s of type %s: %s", obj.GetNamespace(), obj.GetName(), "ComplianceScan", err)
	}
	if !found {
		return "", fmt.Errorf("%s/%s of type %s: has no phase in status", obj.GetNamespace(), h.name, h.kind)
	}
	return phase, nil
}

func (h *ComplianceScanHelper) getCurrentIndex(obj *unstructured.Unstructured) (int64, error) {
	curri, found, err := unstructured.NestedInt64(obj.Object, "status", "currentIndex")
	if err != nil {
		return 0, fmt.Errorf("Unable to get currentIndex of %s/%s of type %s: %s", obj.GetNamespace(), obj.GetName(), h.kind, err)
	}
	if !found {
		return 0, nil
	}
	return curri, nil
}

func (h *ComplianceScanHelper) getResultsStorageRef(obj *unstructured.Unstructured) (map[string]string, error) {
	rs, found, err := unstructured.NestedStringMap(obj.Object, "status", "resultsStorage")
	if err != nil {
		return nil, fmt.Errorf("Unable to get resultsStorage of %s/%s of type %s: %s", obj.GetNamespace(), obj.GetName(), "ComplianceScan", err)
	}
	if !found {
		return nil, fmt.Errorf("%s/%s of type %s: has no resultsStorage in status", obj.GetNamespace(), h.name, h.kind)
	}
	return rs, nil
}

func (h *ComplianceScanHelper) waitForExtractorPod(ns, objName string) error {
	sel := labels.SelectorFromSet(getPVCExtractorPodLabels(objName))
	opts := metav1.ListOptions{
		LabelSelector: sel.String(),
	}
	// retry and ignore errors until timeout
	var lastErr error
	fmt.Fprintf(h.Out, "Fetching raw compliance results for scan '%s'.", h.name)
	timeouterr := wait.Poll(common.RetryInterval, common.Timeout, func() (bool, error) {
		podlist, err := h.kuser.Clientset().CoreV1().Pods(ns).List(context.TODO(), opts)
		lastErr = err
		if err != nil {
			// retry
			return false, nil
		}
		if len(podlist.Items) == 0 {
			// wait for the pod to show up
			return false, nil
		}

		pod := podlist.Items[0]
		if pod.Status.Phase == corev1.PodRunning || pod.Status.Phase == corev1.PodSucceeded {
			return true, nil
		}
		fmt.Fprint(h.Out, ".")
		return false, nil
	})
	fmt.Fprint(h.Out, "\n")

	if timeouterr != nil {
		return fmt.Errorf("The extractor pod wasn't ready before the timeout")
	}

	if lastErr != nil {
		return lastErr
	}
	return nil
}

func getPVCExtractorPodLabels(objName string) map[string]string {
	return map[string]string{
		cmdLabelKey:     "",
		objNameLabelKey: objName,
	}
}

func getPVCExtractorPod(objName, ns, claimName string) *corev1.Pod {
	return &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "raw-result-extractor-",
			Namespace:    ns,
			Labels:       getPVCExtractorPodLabels(objName),
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "pv-extract-pod",
					Image:   "registry.access.redhat.com/ubi8/ubi:latest",
					Command: []string{"sleep", "inf"},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "raw-results-vol",
							ReadOnly:  true,
							MountPath: fmt.Sprintf("/%s", rawResultsMountPath),
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "raw-results-vol",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: claimName,
						},
					},
				},
			},
		},
	}
}
