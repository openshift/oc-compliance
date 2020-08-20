package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/cmd/cp"
	"k8s.io/kubectl/pkg/scheme"
)

const (
	cmpAPIGroup        = "compliance.openshift.io"
	cmpResourceVersion = "v1alpha1"
	cmdLabelKey        = "fetch-compliance-results"
	objNameLabelKey    = "fetch-compliance-results/obj-name"
	retryInterval      = time.Second * 2
	timeout            = time.Minute * 20

	rawResultsMountPath = "raw-results"
)

var (
	namespaceExample = `
	# Fetch from compliancescan
	%[1]s fetch-compliance-results compliancescan

	# Fetch from compliancesuite compliancesuite
	%[1]s fetch-compliance-results

	# Fetch from scansettingbindings
	%[1]s fetch-compliance-results scansettingbindings foo
`

	errNoContext = fmt.Errorf("no context is currently set, use %q to select a new one", "kubectl config use-context <context>")
)

type FCROptions struct {
	configFlags *genericclioptions.ConfigFlags

	dynclient dynamic.Interface

	clientset kubernetes.Interface

	cfg *rest.Config

	namespace string

	outputPath string

	args []string

	helper ObjectHelper

	genericclioptions.IOStreams
}

type ObjectHelper interface {
	Handle() error
}

type ScanSettingsBindingHelper struct {
	opts *FCROptions
	gvk  schema.GroupVersionResource
	name string
}

func NewScanSettingsBindingHelper(opts *FCROptions, name string) *ScanSettingsBindingHelper {
	return &ScanSettingsBindingHelper{
		opts: opts,
		name: name,
		gvk: schema.GroupVersionResource{
			Group:    cmpAPIGroup,
			Version:  cmpResourceVersion,
			Resource: "scansettingbindings",
		},
	}
}

func (h *ScanSettingsBindingHelper) Handle() error {
	return nil
}

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

type ComplianceScanHelper struct {
	opts   *FCROptions
	gvk    schema.GroupVersionResource
	podgvk schema.GroupVersionResource
	kind   string
	name   string
}

func NewComplianceScanHelper(opts *FCROptions, name string) *ComplianceScanHelper {
	return &ComplianceScanHelper{
		opts: opts,
		name: name,
		kind: "ComplianceScan",
		gvk: schema.GroupVersionResource{
			Group:    cmpAPIGroup,
			Version:  cmpResourceVersion,
			Resource: "compliancescans",
		},
		podgvk: schema.GroupVersionResource{
			Group:    "",
			Version:  "v1",
			Resource: "pods",
		},
	}
}

func (h *ComplianceScanHelper) Handle() error {
	// Get target resource
	res, err := h.opts.dynclient.Resource(h.gvk).Namespace(h.opts.namespace).Get(context.TODO(), h.name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("Unable to get resource %s/%s of type %s: %s", h.opts.namespace, h.name, h.kind, err)
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
	extractorPod, err = h.opts.clientset.CoreV1().Pods(rsnamespace).Create(context.TODO(), extractorPod, metav1.CreateOptions{})
	if err != nil && !kerrors.IsAlreadyExists(err) {
		return err
	}

	// wait for extractor pod
	err = h.waitForExtractorPod(rsnamespace, res.GetName())
	if err != nil {
		return err
	}

	cpopts := cp.NewCopyOptions(h.opts.IOStreams)
	cpopts.Namespace = rsnamespace
	cpopts.ClientConfig = h.opts.cfg
	cpopts.Clientset = h.opts.clientset

	podName := extractorPod.GetName()
	path := fmt.Sprintf("%s/%d", rawResultsMountPath, ci)
	cpargs := []string{
		fmt.Sprintf("%s/%s:%s", rsnamespace, podName, path),
		h.opts.outputPath,
	}

	if err = cpopts.Run(cpargs); err != nil {
		return err
	}

	fmt.Printf("The raw compliance results are avaliable in the following directory: %s", h.opts.outputPath)

	var zeroGP int64 = 0
	return h.opts.clientset.CoreV1().Pods(rsnamespace).Delete(context.TODO(), extractorPod.GetName(), metav1.DeleteOptions{
		GracePeriodSeconds: &zeroGP,
	})
}

func (h *ComplianceScanHelper) getScanPhase(obj *unstructured.Unstructured) (string, error) {
	phase, found, err := unstructured.NestedString(obj.Object, "status", "phase")
	if err != nil {
		return "", fmt.Errorf("Unable to get phase of %s/%s of type %s: %s", obj.GetNamespace(), obj.GetName(), "ComplianceScan", err)
	}
	if !found {
		return "", fmt.Errorf("%s/%s of type %s: has no phase in status", h.opts.namespace, h.name, h.kind)
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
		return nil, fmt.Errorf("%s/%s of type %s: has no resultsStorage in status", h.opts.namespace, h.name, h.kind)
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
	fmt.Print("Fetching raw compliance results.")
	timeouterr := wait.Poll(retryInterval, timeout, func() (bool, error) {
		podlist, err := h.opts.clientset.CoreV1().Pods(ns).List(context.TODO(), opts)
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
		fmt.Print(".")
		return false, nil
	})
	fmt.Print("\n")

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

func NewFCROptions(streams genericclioptions.IOStreams) *FCROptions {
	return &FCROptions{
		configFlags: genericclioptions.NewConfigFlags(true),
		IOStreams:   streams,
	}
}

func NewCmdFCR(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewFCROptions(streams)

	cmd := &cobra.Command{
		Use:          "oc fcr [object] [object name] -o [output path]",
		Short:        "",
		Example:      fmt.Sprintf(namespaceExample, "oc"),
		SilenceUsage: true,
		RunE: func(c *cobra.Command, args []string) error {
			if err := o.Complete(c, args); err != nil {
				return err
			}
			if err := o.Validate(); err != nil {
				return err
			}
			if err := o.Run(); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&o.outputPath, "output", "o", ".", "The path where you want to persist the raw results to")
	o.configFlags.AddFlags(cmd.Flags())

	return cmd
}

// Complete sets all information required for updating the current context
func (o *FCROptions) Complete(cmd *cobra.Command, args []string) error {
	o.args = args

	var err error

	o.cfg, err = o.configFlags.ToRESTConfig()
	if err != nil {
		panic(err)
	}

	// NOTE(jaosorior): workaround for https://github.com/kubernetes/client-go/issues/657
	o.cfg.GroupVersion = &schema.GroupVersion{Group: "", Version: "v1"}
	o.cfg.NegotiatedSerializer = scheme.Codecs.WithoutConversion()
	// NOTE(jaosorior): This is needed so the appropriate api path is used.
	o.cfg.APIPath = "/api"

	o.clientset, err = kubernetes.NewForConfig(o.cfg)
	if err != nil {
		panic(err)
	}

	o.dynclient, err = dynamic.NewForConfig(o.cfg)
	if err != nil {
		panic(err)
	}

	rawConfig, err := o.configFlags.ToRawKubeConfigLoader().RawConfig()
	if err != nil {
		return err
	}

	if currentContext, exists := rawConfig.Contexts[rawConfig.CurrentContext]; exists {
		if currentContext.Namespace != "" {
			o.namespace = currentContext.Namespace
		}
	}

	// Takes precedence
	givenNamespace, err := cmd.Flags().GetString("namespace")
	if err != nil {
		return err
	}
	if givenNamespace != "" {
		o.namespace = givenNamespace
	}
	return nil
}

// Validate ensures that all required arguments and flag values are provided
func (o *FCROptions) Validate() error {
	finfo, err := os.Stat(o.outputPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("The directory at path '%s' doesn't exist\n", o.outputPath)
	}

	if !finfo.IsDir() {
		return fmt.Errorf("The output path must be a directory")
	}

	if len(o.args) < 1 {
		return fmt.Errorf("You need to specify at least one object")
	}

	if len(o.args) > 2 {
		return fmt.Errorf("unkown argument(s): %s", o.args[2:])
	}

	var rawobjtype, objname string
	if len(o.args) == 1 {
		objref := strings.Split(o.args[0], "/")
		if len(objref) == 1 {
			return fmt.Errorf("Missing object name")
		}

		if len(objref) > 2 {
			return fmt.Errorf("Malformed reference to object: %s", o.args[0])
		}

		rawobjtype = o.args[0]
		objname = o.args[1]
	} else {
		rawobjtype = o.args[0]
		objname = o.args[1]
	}

	switch rawobjtype {
	case "ScanSettingBindings", "ScanSettingBinding", "scansettingbindings", "scansettingbinding":
		o.helper = NewScanSettingsBindingHelper(o, objname)
	case "ComplianceSuites", "ComplianceSuite", "compliancesuites", "compliancesuite":
		o.helper = NewComplianceSuiteHelper(o, objname)
	case "ComplianceScans", "ComplianceScan", "compliancescans", "compliancescan":
		o.helper = NewComplianceScanHelper(o, objname)
	default:
		return fmt.Errorf("Unkown object type: %s", rawobjtype)
	}

	return nil
}

func (o *FCROptions) Run() error {
	return o.helper.Handle()
}
