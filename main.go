package main

import (
	"os"

	"github.com/alecthomas/kong"
	"github.com/mnixry/edgeone-sls-push/internal/app"
	"github.com/mnixry/edgeone-sls-push/internal/config"
	"github.com/mnixry/edgeone-sls-push/internal/logger"
)

func main() {
	var cli config.CLI
	kong.Parse(&cli,
		kong.UsageOnError(),
	)
	log := logger.New(cli.Log)
	if err := app.Run(cli, log); err != nil {
		log.Fatal().Err(err).Send()
		os.Exit(1)
	}
}
