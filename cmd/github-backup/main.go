package main

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"

	"github.com/JenswBE/github-backup/internal/backup"
	"github.com/JenswBE/github-backup/internal/config"
	"github.com/JenswBE/github-backup/internal/logging"
)

func main() {
	// Parse config
	svcConfig, err := config.ParseConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse config")
	}

	// Parse flags
	verbose := pflag.BoolP("verbose", "v", false, "Enable verbose output")
	pflag.Parse()

	// Apply flags on config
	if !svcConfig.Verbose {
		svcConfig.Verbose = *verbose
	}

	// Setup logging
	logging.Setup(svcConfig.Verbose, svcConfig.Console)

	// Run backup
	if err = backup.Backup(svcConfig); err != nil {
		log.Fatal().Err(err).Msg("Backup of GitHub failed")
	}
}
