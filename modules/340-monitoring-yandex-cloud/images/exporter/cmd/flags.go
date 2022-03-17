package main

import (
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	serviceAccountPath = ""
	apiKey             = ""
	apiKeyFilePath     = ""
	folderID           = ""
	listenAddress      = "127.0.0.1:9000"
	services           = make([]string, 0)
	loggerType         = loggerJSON
	loggerLevel        = int(logrus.InfoLevel)
	autoRenewIAMToken  = 1 * time.Hour
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

	cmd.Flag("folder-id", "Yandex folder id").
		Envar("FOLDER_ID").
		Required().
		StringVar(&folderID)

	cmd.Flag("listen-address", "Listen address for HTTP").
		Envar("LISTEN_ADDRESS").
		Default(listenAddress).
		StringVar(&listenAddress)

	cmd.Flag("logger-type", "Format logs output of a dhctl in different ways.").
		Envar("LOGGER_TYPE").
		Default(loggerType).
		EnumVar(&loggerType, loggerJSON, loggerSimple)

	cmd.Flag("v", "Logger verbosity").
		Envar("LOGGER_LEVEL").
		Default(strconv.Itoa(int(loggerLevel))).
		IntVar(&loggerLevel)

	cmd.Flag("services", "List services for '/metrics' path").
		Envar("HTTP_PORT").
		StringsVar(&services)

	cmd.Flag("auto-renew-iam-token-period", "Period for renew yandex IAM-token for service account").
		Envar("HTTP_PORT").
		Default(autoRenewIAMToken.String()).
		DurationVar(&autoRenewIAMToken)

}
