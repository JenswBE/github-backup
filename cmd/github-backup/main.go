package main

import (
	"time"

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
	ignoreMaxFoldersToDelete := pflag.Bool("ignore-max-folders-to-delete", false, "Ignores MaxFoldersToDelete")
	pflag.Parse()

	// Apply flags on config
	if !svcConfig.Verbose {
		svcConfig.Verbose = *verbose
	}
	if *ignoreMaxFoldersToDelete {
		svcConfig.MaxFoldersToDelete = -1
	}

	// Setup logging
	logging.Setup(svcConfig.Verbose, svcConfig.Console)

	// Run backup
	log.Info().Msg("Starting backup ...")
	start := time.Now()
	err = backup.Backup(svcConfig)
	logger := log.With().Dur("duration_sec", time.Since(start)).Logger()
	if err != nil {
		logger.Fatal().Err(err).Msg("Backup failed")
	}
	logger.Info().Msg("Backup finished")
}
