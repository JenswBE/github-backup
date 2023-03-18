package git

import (
	"fmt"
	"os/exec"

	"github.com/rs/zerolog/log"
)

// Init creates a new mirrored repository in provided local path
func Init(repoURL AuthenticatedURL, localPath string) error {
	cmd := exec.Command("git", "clone", "--mirror", repoURL.String(), localPath)
	output, err := cmd.CombinedOutput()
	logger := log.With().Bytes("output", output).Stringer("command", cmd).Str("path", localPath).Logger()
	if err != nil {
		logger.Error().Msg("Failed to init new mirrored repository")
		return fmt.Errorf("failed to update remote URL of local mirror: %w", err)
	} else {
		logger.Debug().Msg("Successfully initialized a new mirrored repository")
	}
	return nil
}

// Update syncs the local mirror from the remote source
func Update(repoURL AuthenticatedURL, localPath string) error {
	// Ensure PAT is up-to-date
	cmd := exec.Command("git", "remote", "set-url", "origin", repoURL.String())
	cmd.Dir = localPath
	output, err := cmd.CombinedOutput()
	logger := log.With().Bytes("output", output).Stringer("command", cmd).Str("path", localPath).Logger()
	if err != nil {
		logger.Error().Msg("Failed to remote URL of repo")
		return fmt.Errorf("failed to update remote URL of local mirror: %w", err)
	} else {
		logger.Debug().Msg("Successfully updated remote URL of repo")
	}

	// Update mirrored repo
	cmd = exec.Command("git", "remote", "update", "--prune")
	cmd.Dir = localPath
	output, err = cmd.CombinedOutput()
	logger = log.With().Bytes("output", output).Stringer("command", cmd).Str("path", localPath).Logger()
	if err != nil {
		logger.Error().Msg("Failed to update local mirror")
		return fmt.Errorf("failed to update local mirror: %w", err)
	} else {
		logger.Debug().Msg("Successfully updated local mirror")
	}
	return nil
}
