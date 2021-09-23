package logger

/* -----------------------------------------------------------------
 * log.go -
 *
 * Sets logging format for project
 *
 * Author: Caleb Carlson
 *
 * © Copyright 2021 Hewlett Packard Enterprise Development LP
 *
 * ----------------------------------------------------------------- */

import (
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

const projectPath = "rfsf-openapi"

// Automatically called by logrus
func SetupLogging() {
	log.WithFields(log.Fields{"LogLevel": log.GetLevel()}).Info("Logging Initialized")
}

func init() {
	logLevel := os.Getenv("LOG_LEVEL")
	logLevel = strings.ToUpper(logLevel)

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: false,
		ForceColors:   true,
	})
	log.SetReportCaller(false)

	switch logLevel {
	case "TRACE":
		log.SetLevel(log.TraceLevel)
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
	case "INFO":
		log.SetLevel(log.InfoLevel)
	case "WARN":
		log.SetLevel(log.WarnLevel)
	case "ERROR":
		log.SetLevel(log.ErrorLevel)
	case "FATAL":
		log.SetLevel(log.FatalLevel)
	case "PANIC":
		log.SetLevel(log.PanicLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}
}
