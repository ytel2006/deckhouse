/*
Copyright 2022 Flant JSC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package template_tests

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/deckhouse/deckhouse/testing/helm"
)

var _ = Describe("Module :: monitoring-yandex-cloud :: helm template", func() {
	const global = `
discovery:
  d8SpecificNodeCountByRole: {}
modulesImages:
  registry: registry.deckhouse.io/deckhouse/ce
  registryDockercfg: Y2ZnCg==
  registryAddress: registry.deckhouse.io
  registryPath: /deckhouse/ce
  registryScheme: https
  tags:
    common:
      kubeRbacProxy: imagehash
      alpine: tagstring
    monitoringYandexCloud:
      exporter: exporter_image
`
	assertExporterDeploymentAndSecret := func(h *Config) {
		deployment := h.KubernetesResource("Deployment", "d8-monitoring", "yandex-cloud-metrics-exporter")
		Expect(deployment.Exists()).To(BeTrue())

		secret := h.KubernetesResource("Secret", "d8-monitoring", "d8-yandex-metrics-exporter-app-creds")
		Expect(secret.Exists()).To(BeTrue())
		Expect(secret.Field("data.api-key").String()).To(Equal("YXBpLWtleQ=="))
		Expect(secret.Field("data.folder-id").String()).To(Equal("Zm9sZGVyLWlk"))
	}

	hec := SetupHelmConfig("")

	BeforeEach(func() {
		hec.ValuesSetFromYaml("global", global)

		hec.ValuesSet("monitoringYandexCloud.apiKey", "api-key")
		hec.ValuesSet("monitoringYandexCloud.internal.folderID", "folder-id")
	})

	Context("with nat instance", func() {
		BeforeEach(func() {
			hec.ValuesSet("monitoringYandexCloud.internal.natInstanceName", "cluster-nat-instance")
			hec.HelmRender()
		})

		It("Should create deployment with exporter and secret with creds for exporter", func() {
			Expect(hec.RenderError).ShouldNot(HaveOccurred())

			assertExporterDeploymentAndSecret(hec)
		})

		It("Should create ServiceMonitor for export nat-instance-metrics", func() {
			Expect(hec.RenderError).ShouldNot(HaveOccurred())

			monitor := hec.KubernetesResource("ServiceMonitor", "d8-monitoring", "yandex-nat-instance-metrics")
			Expect(monitor.Exists()).To(BeTrue())
		})
	})

	Context("with nat instance, but scraping was disabled from config", func() {
		BeforeEach(func() {
			hec.ValuesSet("monitoringYandexCloud.disableScrapeNATInstanceMetrics", true)
			hec.ValuesSet("monitoringYandexCloud.internal.natInstanceName", "cluster-nat-instance")
			hec.HelmRender()
		})

		It("Should create deployment with exporter and secret with creds for exporter", func() {
			Expect(hec.RenderError).ShouldNot(HaveOccurred())

			assertExporterDeploymentAndSecret(hec)
		})

		It("Should not create ServiceMonitor for export nat-instance-metrics", func() {
			Expect(hec.RenderError).ShouldNot(HaveOccurred())

			deployment := hec.KubernetesResource("ServiceMonitor", "d8-monitoring", "yandex-nat-instance-metrics")
			Expect(deployment.Exists()).To(BeFalse())
		})
	})

	Context("without nat instance", func() {
		BeforeEach(func() {
			hec.HelmRender()
		})

		It("Should create deployment with exporter and secret with creds for exporter", func() {
			Expect(hec.RenderError).ShouldNot(HaveOccurred())

			assertExporterDeploymentAndSecret(hec)
		})

		It("Should not create ServiceMonitor for export nat-instance metrics", func() {
			Expect(hec.RenderError).ShouldNot(HaveOccurred())

			monitor := hec.KubernetesResource("ServiceMonitor", "d8-monitoring", "yandex-nat-instance-metrics")
			Expect(monitor.Exists()).To(BeFalse())
		})
	})
})
