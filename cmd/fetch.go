package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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
	usageExamples = `
  # Fetch from compliancescan
  %[1]s %[2]s compliancescan [resource name] -o [directory]
  
  # Fetch from compliancesuite
  %[1]s %[2]s compliancesuite [resource name] -o [directory]
  
  # Fetch from scansettingbindings
  %[1]s %[2]s scansettingbindings [resource name] -o [directory]
`

	errNoContext = fmt.Errorf("no context is currently set, use %q to select a new one", "kubectl config use-context <context>")
)

func init() {
	fetchRawCmd := NewCmdFCR(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
	rootCmd.AddCommand(fetchRawCmd)
}

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

func NewFCROptions(streams genericclioptions.IOStreams) *FCROptions {
	return &FCROptions{
		configFlags: genericclioptions.NewConfigFlags(true),
		IOStreams:   streams,
	}
}

func NewCmdFCR(streams genericclioptions.IOStreams) *cobra.Command {
	o := NewFCROptions(streams)

	cmd := &cobra.Command{
		Use:   "fcr [object] [object name] -o [output path]",
		Short: "Download raw compliance results",
		Long: `'fcr' stands for 'fetch-compliance-results'.

This command allows you to download the raw results from a
compliance scan to a specified directory.`,
		Example:      fmt.Sprintf(usageExamples, "oc", "fcr"),
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
		return fmt.Errorf("The directory at path '%s' doesn't exist", o.outputPath)
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

		rawobjtype = objref[0]
		objname = objref[1]
	} else {
		rawobjtype = o.args[0]
		objname = o.args[1]
	}

	switch rawobjtype {
	case "ScanSettingBindings", "ScanSettingBinding", "scansettingbindings", "scansettingbinding":
		o.helper = NewScanSettingBindingHelper(o, objname, o.outputPath)
	case "ComplianceSuites", "ComplianceSuite", "compliancesuites", "compliancesuite":
		o.helper = NewComplianceSuiteHelper(o, objname, o.outputPath)
	case "ComplianceScans", "ComplianceScan", "compliancescans", "compliancescan":
		o.helper = NewComplianceScanHelper(o, objname, o.outputPath)
	default:
		return fmt.Errorf("Unkown object type: %s", rawobjtype)
	}

	return nil
}

func (o *FCROptions) Run() error {
	return o.helper.Handle()
}
