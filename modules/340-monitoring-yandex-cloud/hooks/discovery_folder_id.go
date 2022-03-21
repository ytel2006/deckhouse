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
	"encoding/json"
	"fmt"

	"github.com/flant/addon-operator/pkg/module_manager/go_hook"
	"github.com/flant/addon-operator/sdk"
	"github.com/flant/shell-operator/pkg/kube_events_manager/types"
	v1core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/deckhouse/deckhouse/dhctl/pkg/config"
)

var _ = sdk.RegisterFunc(&go_hook.HookConfig{
	OnBeforeHelm: &go_hook.OrderedConfig{Order: 10},
	Kubernetes: []go_hook.KubernetesConfig{
		{
			Name:       "folder_id",
			ApiVersion: "v1",
			Kind:       "Secret",
			NamespaceSelector: &types.NamespaceSelector{
				NameSelector: &types.NameSelector{
					MatchNames: []string{"kube-system"},
				},
			},
			NameSelector: &types.NameSelector{
				MatchNames: []string{"d8-provider-cluster-configuration"},
			},
			FilterFunc: applyFolderID,
		},
	},
}, discoveryFolderId)

func applyFolderID(obj *unstructured.Unstructured) (go_hook.FilterResult, error) {
	var secret = &v1core.Secret{}
	err := sdk.FromUnstructured(obj, secret)
	if err != nil {
		return "", fmt.Errorf("cannot convert secret from unstructured: %v", err)
	}

	clusterConfigurationYAML, ok := secret.Data["cloud-provider-cluster-configuration.yaml"]

	if !ok || len(clusterConfigurationYAML) == 0 {
		return "", fmt.Errorf("cloud-provider-cluster-configuration.yaml not found or empty: %v", err)
	}

	m, err := config.ParseConfigFromData(string(clusterConfigurationYAML))
	if err != nil {
		return "", fmt.Errorf("validate cloud-provider-cluster-configuration.yaml: %v", err)
	}

	var provider map[string]interface{}
	if err := json.Unmarshal(m.ProviderClusterConfig["provider"], &provider); err != nil {
		return "", fmt.Errorf("cannot decode `provider` property from provider cluster configuration: %v", err)
	}

	folderIDRaw, ok := provider["folderID"]
	if !ok {
		return "", fmt.Errorf("folderID not found in provider")
	}

	folderID, ok := folderIDRaw.(string)
	if !ok {
		return "", fmt.Errorf("folderID is not string")
	}

	if folderID == "" {
		return "", fmt.Errorf("folderID is empty")
	}

	return folderID, nil
}

// discoveryFolderId
// There is CM kube-system/d8-cluster-uuid with cluster uuid. Hook must store it to `global.discovery.clusterUUID`.
// Or generate uuid and create CM
func discoveryFolderId(input *go_hook.HookInput) error {
	folderIDSnap := input.Snapshots["folder_id"]
	if len(folderIDSnap) == 0 {
		return fmt.Errorf("yandex provider cloud configuration secret not found")
	}

	folderID := folderIDSnap[0].(string)
	input.Values.Set("monitoringYandexCloud.internal.folderID", folderID)

	return nil
}
