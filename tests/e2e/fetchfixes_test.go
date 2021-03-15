package e2e

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func assertFetchFixesFetchedSomething(dir string) {
	By("Finding remediations")
	dirraw := do("find", dir, "-name", "*.yaml")
	dirs := strings.Split(dirraw, "\n")

	By("Assert we got results")
	Expect(len(dirs)).Should(BeNumerically(">", 0))
}

var _ = Describe("fetch-fixes", func() {
	var dir string
	BeforeEach(func() {
		var tmpErr error
		dir, tmpErr = ioutil.TempDir("", "fetch-fixes")
		Expect(tmpErr).ShouldNot(HaveOccurred())
		By(fmt.Sprintf("Created temporary directory for this test: %s", dir))
	})
	Context("With a remediation", func() {
		var targetRem string

		BeforeEach(func() {
			withCISScan("fetch-fixes-scan")
			rems := oc("get", "complianceremediations",
				"-o", `jsonpath={range .items[:]}{.metadata.name}{"\n"}{end}`)
			remsSlice := strings.Split(rems, "\n")
			idx := rand.Intn(len(remsSlice))
			targetRem = remsSlice[idx]
		}, float64(scanDoneTimeout))

		It("fetches fixes for ComplianceRemediation", func() {
			oc("compliance", "fetch-fixes", "complianceremediation", targetRem, "-o", dir)
		})
	})

	Context("With parsed content", func() {
		It("fetches fixes for profile", func() {
			oc("compliance", "fetch-fixes", "profile", "ocp4-cis", "-o", dir)
		})
	})
})
