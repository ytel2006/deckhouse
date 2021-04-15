package hooks

import (
	"io/ioutil"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/deckhouse/deckhouse/testing/hooks"
)

var _ = Describe("Istio hooks :: multicluster_discovery_api_hosts ::", func() {
	f := HookExecutionConfigInit(`{"istio":{"multicluster":{},"internal":{"clientCertificate":{"cert":"ccc","key":"kkk"}}}}`, "")
	f.RegisterCRD("deckhouse.io", "v1alpha1", "IstioMulticluster", false)

	Context("Empty cluster and minimal settings", func() {
		BeforeEach(func() {
			f.BindingContexts.Set(f.KubeStateSet(``))
			f.RunHook()
		})

		It("Hook must execute successfully", func() {
			Expect(f).To(ExecuteSuccessfully())

			stderrBuff := string(f.Session.Err.Contents())
			Expect(stderrBuff).To(Equal(""))
		})
	})

	Context("Empty cluster, minimal settings and multicluster is enabled", func() {
		BeforeEach(func() {
			f.ValuesSet("istio.multicluster.enabled", true)
			f.BindingContexts.Set(f.KubeStateSet(``))
			f.RunHook()
		})

		It("Hook must execute successfully", func() {
			Expect(f).To(ExecuteSuccessfully())

			stderrBuff := string(f.Session.Err.Contents())
			Expect(stderrBuff).To(Equal(""))
		})
	})

	Context("Proper multiclusters only", func() {
		BeforeEach(func() {
			f.ValuesSet(`istio.multicluster.enabled`, true)
			f.BindingContexts.Set(f.KubeStateSet(`
---
apiVersion: deckhouse.io/v1alpha1
kind: IstioMulticluster
metadata:
  name: proper-multicluster-0
spec:
  metadataEndpoint: "file:///tmp/proper-multicluster-0/"
status: {}
---
apiVersion: deckhouse.io/v1alpha1
kind: IstioMulticluster
metadata:
  name: proper-multicluster-1
spec:
  trustDomain: "p.f1"
  metadataEndpoint: "file:///tmp/proper-multicluster-1/"
status:
  metadataCache:
    apiHost: some-outdated.host
`))
			_ = os.MkdirAll("/tmp/proper-multicluster-0/private", 0755)
			ioutil.WriteFile("/tmp/proper-multicluster-0/private/multicluster-api-host", []byte(`istio-api.0.com`), 0644)
			_ = os.MkdirAll("/tmp/proper-multicluster-1/private", 0755)
			ioutil.WriteFile("/tmp/proper-multicluster-1/private/multicluster-api-host", []byte(`istio-api.1.com`), 0644)

			f.RunHook()
		})

		It("Hook must execute successfully", func() {
			Expect(f).To(ExecuteSuccessfully())

			stderrBuff := string(f.Session.Err.Contents())
			Expect(stderrBuff).To(Equal(""))

			t0, err := time.Parse(time.RFC3339, f.KubernetesGlobalResource("IstioMulticluster", "proper-multicluster-0").Field("status.metadataCache.apiHostLastFetchTimestamp").String())
			Expect(err).ShouldNot(HaveOccurred())
			t1, err := time.Parse(time.RFC3339, f.KubernetesGlobalResource("IstioMulticluster", "proper-multicluster-1").Field("status.metadataCache.apiHostLastFetchTimestamp").String())
			Expect(err).ShouldNot(HaveOccurred())

			Expect(t0).Should(BeTemporally("~", time.Now().UTC(), time.Minute))
			Expect(t1).Should(BeTemporally("~", time.Now().UTC(), time.Minute))

			Expect(f.KubernetesGlobalResource("IstioMulticluster", "proper-multicluster-0").Field("status.metadataCache.apiHost").String()).To(Equal("istio-api.0.com"))
			Expect(f.KubernetesGlobalResource("IstioMulticluster", "proper-multicluster-1").Field("status.metadataCache.apiHost").String()).To(Equal("istio-api.1.com"))
		})
	})

	Context("Improper multicluster", func() {
		BeforeEach(func() {
			f.ValuesSet(`istio.multicluster.enabled`, true)
			f.BindingContexts.Set(f.KubeStateSet(`
---
apiVersion: deckhouse.io/v1alpha1
kind: IstioMulticluster
metadata:
  name: improper-multicluster-0
spec:
  metadataEndpoint: "https://some-improper-hostname-0/metadata/"
`))

			f.RunHook()
		})

		It("Hook must execute successfully with proper warnings", func() {
			Expect(f).To(ExecuteSuccessfully())

			stderrBuff := string(f.Session.Err.Contents())
			Expect(stderrBuff).Should(ContainSubstring(`ERROR: Cannot fetch api host metadata endpoint https://some-improper-hostname-0/metadata/private/multicluster-api-host for IstioMulticluster improper-multicluster-0.`))
		})
	})
})