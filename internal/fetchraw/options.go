package fetchraw

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/kubectl/pkg/scheme"
)

type FetchRawOptions struct {
	ConfigFlags *genericclioptions.ConfigFlags

	dynclient dynamic.Interface

	clientset kubernetes.Interface

	cfg *rest.Config

	namespace string

	OutputPath string

	args []string

	helper ObjectHelper

	genericclioptions.IOStreams
}

type ObjectHelper interface {
	Handle() error
}

// Complete sets all information required for updating the current context
func (o *FetchRawOptions) Complete(cmd *cobra.Command, args []string) error {
	o.args = args

	var err error

	o.cfg, err = o.ConfigFlags.ToRESTConfig()
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

	rawConfig, err := o.ConfigFlags.ToRawKubeConfigLoader().RawConfig()
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
func (o *FetchRawOptions) Validate() error {
	finfo, err := os.Stat(o.OutputPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("The directory at path '%s' doesn't exist", o.OutputPath)
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
		o.helper = NewScanSettingBindingHelper(o, objname, o.OutputPath)
	case "ComplianceSuites", "ComplianceSuite", "compliancesuites", "compliancesuite":
		o.helper = NewComplianceSuiteHelper(o, objname, o.OutputPath)
	case "ComplianceScans", "ComplianceScan", "compliancescans", "compliancescan":
		o.helper = NewComplianceScanHelper(o, objname, o.OutputPath)
	default:
		return fmt.Errorf("Unkown object type: %s", rawobjtype)
	}

	return nil
}

func (o *FetchRawOptions) Run() error {
	return o.helper.Handle()
}
