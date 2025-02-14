/*
Copyright 2021 Flant JSC
Licensed under the Deckhouse Platform Enterprise Edition (EE) license. See https://github.com/deckhouse/deckhouse/blob/main/ee/LICENSE
*/

package hooks

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/flant/addon-operator/pkg/module_manager/go_hook"
	"github.com/flant/addon-operator/sdk"
	"github.com/flant/shell-operator/pkg/kube_events_manager/types"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/deckhouse/deckhouse/dhctl/pkg/config"
)

func applySecretFilter(obj *unstructured.Unstructured) (go_hook.FilterResult, error) {
	secret := &v1.Secret{}
	err := sdk.FromUnstructured(obj, secret)
	if err != nil {
		return nil, err
	}

	ccYaml, ok := secret.Data["cluster-configuration.yaml"]
	if !ok {
		return nil, fmt.Errorf(`"cluster-configuration.yaml" not found in "d8-cluster-configuration" Secret`)
	}

	return ccYaml, nil
}

func applyConfigMapFilter(obj *unstructured.Unstructured) (go_hook.FilterResult, error) {
	secret := &v1.ConfigMap{}
	err := sdk.FromUnstructured(obj, secret)
	if err != nil {
		return nil, err
	}

	return true, nil
}

var (
	_ = sdk.RegisterFunc(&go_hook.HookConfig{
		Queue: "/modules/flant-integration",
		Kubernetes: []go_hook.KubernetesConfig{
			{
				Name:       "clusterConfiguration",
				ApiVersion: "v1",
				Kind:       "Secret",
				NameSelector: &types.NameSelector{
					MatchNames: []string{"d8-cluster-configuration"},
				},
				NamespaceSelector: &types.NamespaceSelector{
					NameSelector: &types.NameSelector{
						MatchNames: []string{"kube-system"},
					},
				},
				FilterFunc: applySecretFilter,
			},
			{
				Name:       "migrationConfigMap",
				ApiVersion: "v1",
				Kind:       "ConfigMap",
				NameSelector: &types.NameSelector{
					MatchNames: []string{"d8-migrate-cluster-kubernetes-version"},
				},
				NamespaceSelector: &types.NamespaceSelector{
					NameSelector: &types.NameSelector{
						MatchNames: []string{"d8-flant-integration"},
					},
				},
				FilterFunc: applyConfigMapFilter,
			},
		},
	}, migrateClusterKubernetesVersion)
)

func migrateClusterKubernetesVersion(input *go_hook.HookInput) error {
	cm, ok := input.Snapshots["migrationConfigMap"]
	if ok && len(cm) == 1 && cm[0].(bool) {
		input.LogEntry.Info(`find d8-flant-integration/d8-migrate-cluster-kubernetes-version configMap, migration was done, skipping`)
		return nil
	}

	currentConfig, ok := input.Snapshots["clusterConfiguration"]
	if !ok || len(currentConfig) == 0 {
		input.LogEntry.Info(`cannot find kube-system/d8-cluster-configuration secret, skipping`)
		return nil
	}

	// FilterResult is a YAML encoded as a JSON string. Unmarshal it.
	configYamlBytes := currentConfig[0].([]byte)

	var metaConfig *config.MetaConfig
	metaConfig, err := config.ParseConfigFromData(string(configYamlBytes))
	if err != nil {
		return err
	}

	var kubernetesVersionFromMetaConfig string
	err = json.Unmarshal(metaConfig.ClusterConfig["kubernetesVersion"], &kubernetesVersionFromMetaConfig)
	if err != nil {
		return err
	}

	if kubernetesVersionFromMetaConfig != config.DefaultKubernetesVersion {
		// No need to patch secret
		return nil
	}

	b, err := json.Marshal("Automatic")
	if err != nil {
		return err
	}
	metaConfig.ClusterConfig["kubernetesVersion"] = b

	c, err := metaConfig.ClusterConfigYAML()
	if err != nil {
		return err
	}

	encoded := base64.StdEncoding.EncodeToString(c)
	patch := map[string]interface{}{
		"data": map[string]interface{}{
			"cluster-configuration.yaml": encoded,
		},
	}
	input.PatchCollector.MergePatch(patch, "v1", "Secret", "kube-system", "d8-cluster-configuration")

	migrationIsDone := &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "d8-migrate-cluster-kubernetes-version",
			Namespace: "d8-flant-integration",
			Labels: map[string]string{
				"heritage": "deckhouse",
				"module":   "flant-integration",
			},
		},
	}

	input.PatchCollector.Create(migrationIsDone)
	return nil
}
