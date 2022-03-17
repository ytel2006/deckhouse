package main

import (
	log "github.com/sirupsen/logrus"
)

const (
	loggerSimple = "simple"
	loggerJSON   = "json"
)

func initLogger() *log.Entry {
	var formatter log.Formatter = &log.TextFormatter{
		DisableColors:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	}
	if loggerType == loggerJSON {
		formatter = &log.JSONFormatter{}
	}

	log.SetFormatter(formatter)
	log.SetLevel(log.Level(loggerLevel))

	return log.NewEntry(log.New())
}
