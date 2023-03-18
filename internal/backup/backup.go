package backup

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
	log.Debug().Str("backup_path", svcConfig.BackupPath).Msg("Ensuring backup path exists ...")
	if err = os.MkdirAll(svcConfig.BackupPath, 0o700); err != nil {
		return fmt.Errorf("failed to ensure backup path %s exists: %w", svcConfig.BackupPath, err)
	}

	// List folders in backup path
	log.Debug().Str("backup_path", svcConfig.BackupPath).Msg("Listing local folders in backup path ...")
	localFolders, err := listFolders(svcConfig.BackupPath)
	if err != nil {
		return fmt.Errorf("failed to list folders in backup path %s before syncing: %w", svcConfig.BackupPath, err)
	}
	log.Debug().Strs("local_folders", localFolders).Msg("Discovered local folders")
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
		log.Debug().Str("repo", r.GetName()).Str("clone_url", cloneURL).Msg("Backup repo ...")
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
			log.Debug().Str("repo_dir", repoDir).Str("clone_url", cloneURL).Msg("Repo dir not found, initializing a new local folder ...")
			if err = git.Init(authURL, repoDir); err != nil {
				return fmt.Errorf("failed to init new local repo: %w", err)
			}
		} else {
			// Repo dir exists => Update
			log.Debug().Str("repo_dir", repoDir).Str("clone_url", cloneURL).Msg("Repo dir found, updating existing folder ...")
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
		log.Debug().Func(func(e *zerolog.Event) { e.Strs("local_folders", lo.Keys(localFoldersToDelete)) }).Msg("Removing redundant folders ...")
		if svcConfig.MaxFoldersToDelete >= 0 && len(localFoldersToDelete) > svcConfig.MaxFoldersToDelete {
			localFoldersToDeleteList := lo.Keys(localFoldersToDelete)
			log.Error().
				Int("folder_count", len(localFoldersToDeleteList)).
				Int("max_count", svcConfig.MaxFoldersToDelete).
				Strs("folders", localFoldersToDeleteList).
				Msg("Too many folders found to remove")
			return fmt.Errorf("%d redundant folder(s) found, but max is %d", len(localFoldersToDeleteList), svcConfig.MaxFoldersToDelete)
		}
		for f := range localFoldersToDelete {
			rmPath := filepath.Join(svcConfig.BackupPath, f)
			log.Debug().Str("folder", rmPath).Msg("Removing redundant folder ...")
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
		log.Error().
			Int("folder_count", len(folders)).
			Int("repo_count", len(repos)).
			Strs("missing_folders", missingFolders).
			Strs("redundant_folders", redundantFolders).
			Msg("Mismatch in local folders and remote repositories")
		return fmt.Errorf("mismatch in local folders and remote repositories")
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
