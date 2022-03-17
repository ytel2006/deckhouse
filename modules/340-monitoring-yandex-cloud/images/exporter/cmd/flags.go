package main

import "github.com/alecthomas/kingpin"

var (
	serviceAccountPath = ""
	apiKey             = ""
	apiKeyFilePath     = ""
	folderID           = ""
	listenAddress      = "127.0.0.1:9000"
	services           = make([]string, 0)
	loggerType         = loggerJSON
)

func flags(cmd *kingpin.Application) {
	cmd.Flag("service-account-path", "Path to service account file").
		Envar("SERVICE_ACCOUNT_PATH").
		StringVar(&serviceAccountPath)

	cmd.Flag("api-key", "API key for service account").
		Envar("API_KEY").
		StringVar(&apiKey)

	cmd.Flag("api-key-file", "API key file path for service account").
		Envar("API_KEY_PATH").
		StringVar(&apiKeyFilePath)

	cmd.Flag("listen-address", "Listen address for HTTP").
		Envar("LISTEN_ADDRESS").
		Default(listenAddress).
		StringVar(&listenAddress)

	cmd.Flag("services", "List services for '/metrics' path").
		Envar("HTTP_PORT").
		StringsVar(&services)

	cmd.Flag("folder-id", "Yandex folder id").
		Envar("FOLDER_ID").
		Required().
		StringVar(&folderID)

	cmd.Flag("prefix", "Metrics prefix").
		Envar("METRICS_PREFIX").
		Required().
		StringVar(&folderID)

	cmd.Flag("logger-type", "Format logs output of a dhctl in different ways.").
		Envar("LOGGER_TYPE").
		Default(loggerType).
		EnumVar(&loggerType, loggerJSON, loggerSimple)
}
