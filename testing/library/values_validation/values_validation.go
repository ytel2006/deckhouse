package values_validation

import (
	"fmt"
	"path/filepath"

	"github.com/flant/addon-operator/pkg/module_manager"
	"github.com/flant/addon-operator/pkg/utils"
	"github.com/flant/addon-operator/pkg/values/validation"
	"sigs.k8s.io/yaml"
)

func LoadOpenAPISchemas(validator *validation.ValuesValidator, moduleName, modulePath string) error {
	openAPIDir := filepath.Join("/deckhouse", "global-hooks", "openapi")
	configBytes, valuesBytes, err := module_manager.ReadOpenAPIFiles(openAPIDir)
	if err != nil {
		return fmt.Errorf("read global openAPI schemas: %v", err)
	}
	err = validator.SchemaStorage.AddGlobalValuesSchemas(configBytes, valuesBytes)
	if err != nil {
		return fmt.Errorf("parse global openAPI schemas: %v", err)
	}

	if moduleName == "" || modulePath == "" {
		return nil
	}

	valuesKey := utils.ModuleNameToValuesKey(moduleName)
	openAPIPath := filepath.Join(modulePath, "openapi")
	configBytes, valuesBytes, err = module_manager.ReadOpenAPIFiles(openAPIPath)
	if err != nil {
		return fmt.Errorf("module '%s' read openAPI schemas: %v", moduleName, err)
	}

	err = validator.SchemaStorage.AddModuleValuesSchemas(valuesKey, configBytes, valuesBytes)
	if err != nil {
		return fmt.Errorf("parse global openAPI schemas: %v", err)
	}

	return nil
}

// ValidateValues is an adapter between JSONRepr and Values
// TODO There was validating with config-values.yaml schema and ignoring of "internal" key. It seems not needed after x-extend implementation.
func ValidateValues(validator *validation.ValuesValidator, moduleName, values string) error {
	var obj map[string]interface{}
	err := yaml.Unmarshal([]byte(values), &obj)
	if err != nil {
		return err
	}

	err = validator.ValidateGlobalValues(obj)
	if err != nil {
		return err
	}

	valuesKey := utils.ModuleNameToValuesKey(moduleName)
	err = validator.ValidateModuleValues(valuesKey, obj)
	if err != nil {
		return err
	}
	return nil
}

// ValidateValues is an adapter between JSONRepr and Values
func ValidateHelmValues(validator *validation.ValuesValidator, moduleName, values string) error {
	var obj map[string]interface{}
	err := yaml.Unmarshal([]byte(values), &obj)
	if err != nil {
		return err
	}

	err = validator.ValidateGlobalValues(obj)
	if err != nil {
		return err
	}

	valuesKey := utils.ModuleNameToValuesKey(moduleName)
	err = validator.ValidateModuleHelmValues(valuesKey, obj)
	if err != nil {
		return err
	}
	return nil
}