// Copyright 2022 Flant JSC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package hooks

import (
	"encoding/base64"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/deckhouse/deckhouse/testing/hooks"
)

var _ = Describe("monitoring-yandex-cloud :: discovery folder id ::", func() {
	const (
		initValuesString = `
global:
  discovery: {}
monitoringYandexCloud:
  apiKey: apiKeyTest
  internal: {}
`
		// correct cc
		stateCorrect = `
apiVersion: deckhouse.io/v1
existingNetworkID: enpma5uvcfbkuac1i1jb
kind: YandexClusterConfiguration
layout: WithNATInstance
masterNodeGroup:
  instanceClass:
    cores: 2
    imageID: test
    memory: 4096
  replicas: 1
nodeNetworkCIDR: 10.231.0.0/22
provider:
  cloudID: test
  folderID: yandexFolderID
  serviceAccountJSON: |-
    {
      "id": "test"
    }
sshPublicKey: ssh-rsa test
withNATInstance:
  internalSubnetID: test
  natInstanceExternalAddress: 84.201.160.148
nodeNetworkCIDR: 84.201.160.148/31
sshPublicKey: ssh-rsa AAAAAbbbb
`
	)

	secretState := func(cnf string) string {
		return fmt.Sprintf(`
apiVersion: v1
kind: Secret
metadata:
  name: d8-cluster-configuration
  namespace: kube-system
data:
  "cloud-provider-cluster-configuration.yaml": %s
`, base64.StdEncoding.EncodeToString([]byte(cnf)))
	}

	f := HookExecutionConfigInit(initValuesString, `
monitoringYandexCloud:
  apiKey: apiKeyTest
`)

	Context("Cluster has empty state", func() {
		BeforeEach(func() {
			f.BindingContexts.Set(f.KubeStateSet(``))
			f.RunHook()
		})

		It("Hook should fail with errors", func() {
			Expect(f).To(Not(ExecuteSuccessfully()))

			Expect(f.GoHookError.Error()).Should(ContainSubstring("yandex provider cloud configuration secret not found"))
		})
	})

	Context("provider cluster configuration exists", func() {
		BeforeEach(func() {
			f.BindingContexts.Set(f.KubeStateSet(secretState(stateCorrect)))
			f.RunHook()
		})

		It("Hook should set monitoringYandexCloud.internal.folderId", func() {
			Expect(f).To(ExecuteSuccessfully())

			folderID := f.ValuesGet(`monitoringYandexCloud.internal.folderID`).String()

			Expect(folderID).To(Equal("yandexFolderID"))
		})
	})
})
