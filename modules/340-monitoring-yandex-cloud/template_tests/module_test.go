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

		It("Should create desired objects", func() {
			Expect(hec.RenderError).ShouldNot(HaveOccurred())

			deployment := hec.KubernetesResource("Deployment", "d8-monitoring", "yandex-cloud-metrics-exporter")
			Expect(deployment.Exists()).To(BeTrue())

			secret := hec.KubernetesResource("Secret", "d8-monitoring", "yandex-cloud-metrics-exporter")
			Expect(secret.Exists()).To(BeTrue())
			Expect(secret.Field("data.api-key").String()).To(Equal("YXBpLWtleQ=="))
			Expect(secret.Field("data.folder-id").String()).To(Equal("Y2x1c3Rlci1uYXQtaW5zdGFuY2U="))
		})
	})
})
