# GitHub Backup

Automated backups of GitHub repo's using Git mirroring.

## Links

- GitHub: https://github.com/JenswBE/github-backup
- DockerHub: https://hub.docker.com/r/jenswbe/github-backup

## Configuration

GitHub Backup can be configured in 2 ways:

1. Create a file called `config.yml` in the same folder or the parent folder of the binary. See `config.yml` for an example.
2. Set environment variables

If both are defined, the environment variables take precedence.

| Config key             | Env variable                 | Description                                                                        | Default value |
| ---------------------- | ---------------------------- | ---------------------------------------------------------------------------------- | ------------- |
| Username               | GHB_USERNAME                 | GitHub username                                                                    |               |
| PersonalAccessToken    | GHB_PERSONAL_ACCESS_TOKEN    | [GitHub PAT](https://github.com/settings/tokens)                                   |               |
|                        |                              | (Fine-grained with `All repo's` and `Contents` set to `Read-only`)                 |               |
| Verbose                | GHB_VERBOSE                  | Enable verbose logging. Logs sensitive values!                                     | false         |
| Console                | GHB_CONSOLE                  | Enable console logging (default is JSON).                                          | false         |
| BackupPath             | GHB_BACKUP_PATH              | Path to backup repo's to                                                           | ./backup      |
| RemoveRedundantFolders | GHB_REMOVE_REDUNDANT_FOLDERS | Remove local directories for which no repo is found                                | false         |
| MaxFoldersToDelete     | GHB_MAX_FOLDERS_TO_DELETE    | Maximum allowed number of directories to delete.                                   | 3             |
|                        |                              | If more to be deleted, program will return an error without deleting any folder.   |               |
|                        |                              | Can be ignored with --ignore-max-folders-to-delete or by setting a negative value. |               |
