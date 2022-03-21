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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/deckhouse/deckhouse/dhctl/pkg/config"
	"github.com/deckhouse/deckhouse/go_lib/hooks/cluster_configuration"
)

var _ = cluster_configuration.RegisterHook(func(input *go_hook.HookInput, metaCfg *config.MetaConfig, providerDiscoveryData *unstructured.Unstructured, secretFound bool) error {
	if !secretFound {
		return fmt.Errorf("yandex provider cloud configuration secret not found")
	}

	var provider map[string]interface{}
	if err := json.Unmarshal(metaCfg.ProviderClusterConfig["provider"], &provider); err != nil {
		return fmt.Errorf("cannot decode `provider` property from provider cluster configuration: %v", err)
	}

	folderIDRaw, ok := provider["folderID"]
	if !ok {
		return fmt.Errorf("folderID not found in provider")
	}

	folderID, ok := folderIDRaw.(string)
	if !ok {
		return fmt.Errorf("folderID is not string")
	}

	if folderID == "" {
		return fmt.Errorf("folderID is empty")
	}

	natInstanceName, found, err := unstructured.NestedString(providerDiscoveryData.Object, "natInstance", "name")
	if err != nil {
		return fmt.Errorf("nat instance name error %v", err)
	}

	if !found {
		natInstanceName = ""
	}

	input.Values.Set("monitoringYandexCloud.internal.folderID", folderID)
	input.Values.Set("monitoringYandexCloud.internal.natInstanceName", natInstanceName)

	return nil
})
