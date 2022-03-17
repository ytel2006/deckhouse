package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"exporter/internal/yandex"
)

func openFile(path string) (*os.File, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	return os.Open(absPath)
}

func initAPI(api *yandex.CloudApi) error {
	if serviceAccountPath != "" {
		saFile, err := openFile(serviceAccountPath)
		if err != nil {
			return err
		}
		defer saFile.Close()

		return api.InitWithServiceAccount(saFile)
	}

	if apiKeyFilePath != "" {
		apiKeyFile, err := openFile(apiKeyFilePath)
		if err != nil {
			return err
		}

		defer apiKeyFile.Close()

		apiKeyBytes, err := io.ReadAll(apiKeyFile)
		if err != nil {
			return err
		}

		api.InitWithAPIKey(string(apiKeyBytes))
		return nil
	}

	if apiKey != "" {
		api.InitWithAPIKey(apiKey)
		return nil
	}

	return fmt.Errorf("should pass path to service account, pass to file contains APIKey or pass APIKey")
}
