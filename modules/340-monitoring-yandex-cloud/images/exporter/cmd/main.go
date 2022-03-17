package main

import (
	"exporter/internal/server"

	"exporter/internal/yandex"
)

func main() {
	kpApp := kingpin.New("yandex cloud metrics exporter", "A tool for export metrics from yandex cloud in prometheus format")
	kpApp.HelpFlag.Short('h')

	flags(kpApp)

	kpApp.Action(func(context *kingpin.ParseContext) error {
		logger := initLogger()

		stopCh := make(chan struct{}, 1)

		yandexAPI := yandex.NewCloudAPI(logger, folderID, stopCh)
		if err := initAPI(yandexAPI); err != nil {
			return err
		}

		return server.New(logger, yandexAPI).Run(listenAddress, stopCh)
	})

}
