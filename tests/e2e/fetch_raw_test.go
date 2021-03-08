package e2e

import (
	"io/ioutil"
	"os"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const ScanDoneTimeout = 5 * time.Minute

var _ = Describe("fetch-raw", func() {
	Context("With a pre-existing profile being scanned", func() {
		var cwd string

		BeforeEach(func() {
			withCISScan("fetch-raw-scan")
			var tmpErr error
			cwd, tmpErr = ioutil.TempDir("", "oc-compliance-fetch-raw")
			Expect(tmpErr).ShouldNot(HaveOccurred())
		}, float64(ScanDoneTimeout))

		assertFilesOutput := func(dirs []string) {
			By("Assert we got results")
			Expect(len(dirs)).Should(BeNumerically(">", 0))
			for _, dir := range dirs {
				Expect(dir).ToNot(BeEmpty())
			}
		}

		assertFetchRawWorks := func(objtype, objname, wdir string) {
			By("Calling oc compliance fetch-raw")
			oc("compliance", "fetch-raw", objtype, objname, "-o", wdir)

			By("Getting items from scan")
			dirraw := do("ls", wdir)
			dirs := strings.Split(dirraw, "\n")
			assertFilesOutput(dirs)
		}

		assertFetchRawWithHTMLWorks := func(objtype, objname, wdir string) {
			By("Calling oc compliance fetch-raw with --html flag")
			oc("compliance", "fetch-raw", objtype, objname, "-o", wdir)

			By("Getting HTML files from scan")
			dirraw := do("find", wdir, "-name", "*.html")
			dirs := strings.Split(dirraw, "\n")
			assertFilesOutput(dirs)
		}

		When("Fetching objects to directories", func() {
			var dir string

			BeforeEach(func() {
				var tmpErr error
				dir, tmpErr = ioutil.TempDir(cwd, "fetching-obj")
				Expect(tmpErr).ShouldNot(HaveOccurred())
			})

			AfterEach(func() {
				os.RemoveAll(dir)
			})

			When("using ScanSettingBinding", func() {
				It("Fetches the results to the appropriate directory", func() {
					assertFetchRawWorks("scansettingbinding", "fetch-raw-scan", dir)
				})

				It("Fetches the HTML results to the appropriate directory", func() {
					assertFetchRawWithHTMLWorks("scansettingbinding", "fetch-raw-scan", dir)
				})
			})

			When("using ComplianceSuite", func() {
				It("Fetches the results to the appropriate directory", func() {
					assertFetchRawWorks("compliancesuite", "fetch-raw-scan", dir)
				})

				It("Fetches the HTML results to the appropriate directory", func() {
					assertFetchRawWithHTMLWorks("compliancesuite", "fetch-raw-scan", dir)
				})
			})

			When("using ComplianceScan", func() {
				It("Fetches the results to the appropriate directory", func() {
					assertFetchRawWorks("compliancescan", "ocp4-cis", dir)
				})

				It("Fetches the HTML results to the appropriate directory", func() {
					assertFetchRawWithHTMLWorks("compliancescan", "ocp4-cis", dir)
				})
			})
		})

	})
})
