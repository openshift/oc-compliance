package e2e

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const ScanDoneTimeout = 5 * time.Minute

var _ = Describe("FetchRaw", func() {
	Context("With a pre-existing profile being scanned", func() {
		var cwd string

		BeforeEach(func() {
			withCISScan("fetch-raw-scan")
			var tmpErr error
			cwd, tmpErr = ioutil.TempDir("", "oc-compliance-fetch-raw-test")
			Expect(tmpErr).ShouldNot(HaveOccurred())
		}, float64(ScanDoneTimeout))

		Context("ScanSettingBinding", func() {
			var dir string

			BeforeEach(func() {
				dir = filepath.Join(cwd, "ssb")
				os.Mkdir(dir, 0700)
			})

			It("Fetches the results to the appropriate directory", func() {
				By("Calling oc compliance fetch-raw")
				oc("compliance", "fetch-raw", "scansettingbinding", "fetch-raw-scan", "-o", dir)
				By("Getting items from scan")
				dirraw := do("ls", dir)
				dirs := strings.Split(dirraw, "\n")
				Expect(len(dir)).Should(BeNumerically(">", 0))
				for _, dir := range dirs {
					Expect(dir).ToNot(BeEmpty())
				}
			})
		})
	})
})
