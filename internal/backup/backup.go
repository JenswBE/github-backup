package backup

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/samber/lo"

	"github.com/JenswBE/github-backup/internal/config"
	"github.com/JenswBE/github-backup/internal/git"
	"github.com/JenswBE/github-backup/internal/github"
)

func Backup(svcConfig *config.Config) error {
	// List repo's
	ctx := context.Background()
	repos, err := github.ListRepos(ctx, svcConfig.PersonalAccessToken)
	if err != nil {
		return fmt.Errorf("failed to list GitHub repos: %w", err)
	}

	// Ensure backup path exists
	slog.Debug("Ensuring backup path exists ...", "backup_path", svcConfig.BackupPath)
	if err = os.MkdirAll(svcConfig.BackupPath, 0o700); err != nil {
		return fmt.Errorf("failed to ensure backup path %s exists: %w", svcConfig.BackupPath, err)
	}

	// List folders in backup path
	slog.Debug("Listing local folders in backup path ...", "backup_path", svcConfig.BackupPath)
	localFolders, err := listFolders(svcConfig.BackupPath)
	if err != nil {
		return fmt.Errorf("failed to list folders in backup path %s before syncing: %w", svcConfig.BackupPath, err)
	}
	slog.Debug("Discovered local folders", "local_folders", localFolders)
	localFoldersToDelete := lo.SliceToMap(localFolders, func(f string) (string, bool) { return f, true })

	// Backup all repo's
	for i, r := range repos {
		// Check input
		if r == nil {
			return fmt.Errorf("nil repo received on index %d", i)
		}
		if r.GetName() == "" {
			return fmt.Errorf("repo received without name on index %d", i)
		}

		// Backup repo
		repoName := r.GetName()
		cloneURL := r.GetCloneURL()
		slog.Debug("Backup repo ...", "repo", r.GetName(), "clone_url", cloneURL)
		authURL, err := git.GetAuthenticatedURL(cloneURL, svcConfig.Username, svcConfig.PersonalAccessToken)
		if err != nil {
			return fmt.Errorf("failed to get authenticated URL: %w", err)
		}
		repoDir := filepath.Join(svcConfig.BackupPath, repoName)
		repoDirExists, err := pathExists(repoDir)
		if err != nil {
			return fmt.Errorf("failed to check if directory for repo %s already exists: %w", repoName, err)
		}
		if !repoDirExists {
			// Repo dir not found => Init new repo
			slog.Debug("Repo dir not found, initializing a new local folder ...", "repo_dir", repoDir, "clone_url", cloneURL)
			if err = git.Init(authURL, repoDir); err != nil {
				return fmt.Errorf("failed to init new local repo: %w", err)
			}
		} else {
			// Repo dir exists => Update
			slog.Debug("Repo dir found, updating existing folder ...", "repo_dir", repoDir, "clone_url", cloneURL)
			if err = git.Update(authURL, repoDir); err != nil {
				return fmt.Errorf("failed to update local repo: %w", err)
			}

			// Keep local folder
			delete(localFoldersToDelete, repoName)
		}
	}

	// If RemoveRedundantFolders is disabled, no further actions required
	if !svcConfig.RemoveRedundantFolders {
		return nil
	}

	// Delete redundant folders
	if len(localFoldersToDelete) > 0 {
		slog.Debug("Removing redundant folders ...", "local_folders", lo.Keys(localFoldersToDelete))
		if svcConfig.MaxFoldersToDelete >= 0 && len(localFoldersToDelete) > svcConfig.MaxFoldersToDelete {
			localFoldersToDeleteList := lo.Keys(localFoldersToDelete)
			slog.Error("Too many folders found to remove",
				"folder_count", len(localFoldersToDeleteList),
				"max_count", svcConfig.MaxFoldersToDelete,
				"folders", localFoldersToDeleteList,
			)
			return fmt.Errorf("%d redundant folder(s) found, but max is %d", len(localFoldersToDeleteList), svcConfig.MaxFoldersToDelete)
		}
		for f := range localFoldersToDelete {
			rmPath := filepath.Join(svcConfig.BackupPath, f)
			slog.Debug("Removing redundant folder ...", "folder", rmPath)
			if err = os.RemoveAll(rmPath); err != nil {
				return fmt.Errorf("failed to remove redundant folder %s: %w", rmPath, err)
			}
		}
	}

	// Validate remaining folder count matches repo count
	folders, err := listFolders(svcConfig.BackupPath)
	if err != nil {
		return fmt.Errorf("failed to list folders in backup path %s after syncing: %w", svcConfig.BackupPath, err)
	}
	if len(folders) != len(repos) {
		repoNames := github.ExtractRepoNames(repos)
		redundantFolders, missingFolders := lo.Difference(folders, repoNames)
		slog.Error("Mismatch in local folders and remote repositories",
			"folder_count", len(folders),
			"repo_count", len(repos),
			"missing_folders", missingFolders,
			"redundant_folders", redundantFolders,
		)
		return errors.New("mismatch in local folders and remote repositories")
	}

	return nil
}

// listFolders list all base names of folders in specified path.
// Note: Only the name is included, not a relative nor an absolute path.
func listFolders(path string) ([]string, error) {
	folders, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to list contents of folder %s: %w", path, err)
	}
	foldersList := lo.FilterMap(folders, func(f fs.DirEntry, _ int) (string, bool) { return f.Name(), f.IsDir() })
	return foldersList, nil
}

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if path %s exists: %w", path, err)
	}
	return true, nil
}
