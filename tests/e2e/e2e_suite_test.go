package e2e

import (
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestE2e(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2e Suite")
}

var _ = BeforeSuite(func() {
	By("Deploying the Compliance Operator")
	ocApplyFromString(`---
apiVersion: v1
kind: Namespace
metadata:
  name: openshift-compliance
`)

	ocApplyFromString(`---
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: compliance-operator
  namespace: openshift-marketplace
spec:
  displayName: Compliance Operator Upstream
  publisher: github.com/openshift/compliance-operator
  sourceType: grpc
  image: quay.io/compliance-operator/compliance-operator-index:latest

`)

	ocApplyFromString(`---
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: compliance-operator-sub
  namespace: openshift-compliance
spec:
  channel: alpha
  name: compliance-operator
  source: compliance-operator
  sourceNamespace: openshift-marketplace
`)

	ocApplyFromString(`---
apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  name: compliance-operator
  namespace: openshift-compliance
spec:
  targetNamespaces:
  - openshift-compliance
`)

	By("Switching context to compliance-operator NS")
	oc("project", "openshift-compliance")
	time.Sleep(30 * time.Second)

	By("Waiting for all install plans")
	ipraw := oc("get", "installplans", "-o", `jsonpath={range .items[:]}{.metadata.name}{"\n"}{end}`)
	ips := strings.Split(ipraw, "\n")

	for _, ip := range ips {
		ocWaitFor("condition=installed", "installplan", ip)
	}

	By("Waiting for Compliance Operator")

	ocWaitFor("condition=available", "deployment", "compliance-operator")
})

var _ = AfterSuite(func() {
})
