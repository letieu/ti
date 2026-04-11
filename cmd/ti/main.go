package main

import (
	"os"

	"github.com/letieu/ti/internal/cli"
	"github.com/letieu/ti/internal/logger"
)

func main() {
	logLevel := os.Getenv("TI_LOG_LEVEL")
	if logLevel == "" {
		logLevel = "error"
	}

	logFormat := os.Getenv("TI_LOG_FORMAT")
	if logFormat == "json" {
		logger.InitJSON(logger.LogLevel(logLevel))
	} else {
		logger.Init(logger.LogLevel(logLevel))
	}

	c := cli.New()
	c.Run()
}
