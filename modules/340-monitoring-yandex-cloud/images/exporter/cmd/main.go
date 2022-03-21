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

package main

import (
	"os"

	"github.com/sirupsen/logrus"

	"gopkg.in/alecthomas/kingpin.v2"

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

		yandexAPI := yandex.NewCloudAPI(logger, folderID, stopCh).
			WithAutoRenewPeriod(autoRenewIAMToken)

		if err := initAPI(yandexAPI); err != nil {
			return err
		}

		return server.New(logger, yandexAPI, services).Run(listenAddress, stopCh)
	})

	_, err := kpApp.Parse(os.Args[1:])
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
}
