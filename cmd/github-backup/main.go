package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/spf13/pflag"

	"github.com/JenswBE/github-backup/internal/backup"
	"github.com/JenswBE/github-backup/internal/config"
	"github.com/JenswBE/github-backup/internal/logging"
)

func slogFatal(msg string, args ...any) {
	slog.Error(msg, args...)
	os.Exit(1)
}

func main() {
	// Parse config
	svcConfig, err := config.ParseConfig()
	if err != nil {
		slogFatal("Failed to parse config", "error", err)
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
	slog.Info("Starting backup ...")
	start := time.Now()
	err = backup.Backup(svcConfig)
	if err != nil {
		slogFatal("Backup failed", "error", err, "duration_sec", time.Since(start).Seconds())
	}
	slog.Info("Backup finished", "duration_sec", time.Since(start).Seconds())
}
