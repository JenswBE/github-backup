package config

import (
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type Config struct {
	// GitHub username
	Username string
	// GitHub PAT
	PersonalAccessToken string
	// Enable verbose logging
	Verbose bool
	// Enable console logging
	Console bool
	// Path to store the mirrors
	MirrorPath string
	// Maximum number of folders we should delete before considering an error
	MaxFoldersToDelete uint
	// Removes folders for which no matching repo was found
	RemoveRedundantFolders bool
}

func ParseConfig() (*Config, error) {
	// Set defaults
	viper.SetDefault("MirrorPath", ".")
	viper.SetDefault("MaxFoldersToDelete", 3)

	// Parse config file
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("..")
	viper.AddConfigPath("./configs")
	err := viper.ReadInConfig()
	if err != nil {
		var configNotFoundErr viper.ConfigFileNotFoundError
		if !errors.As(err, &configNotFoundErr) {
			return nil, fmt.Errorf("failed reading config file: %w", configNotFoundErr)
		}
		log.Warn().Err(err).Msg("No config file found, expecting configuration through ENV variables")
	}

	// Bind ENV variables
	err = bindEnvs([]envBinding{
		{"Username", "GHB_USERNAME"},
		{"PersonalAccessToken", "GHB_PERSONAL_ACCESS_TOKEN"},
		{"Verbose", "GHB_VERBOSE"},
		{"Console", "GHB_CONSOLE"},
		{"MirrorPath", "GHB_MIRROR_PATH"},
		{"MaxFoldersToDelete", "GHB_MAX_FOLDERS_TO_DELETE"},
	})
	if err != nil {
		return nil, err
	}

	// Unmarshal to Config struct
	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %w", err)
	}
	return &config, nil
}

type envBinding struct {
	configPath string
	envName    string
}

func bindEnvs(bindings []envBinding) error {
	for _, binding := range bindings {
		err := viper.BindEnv(binding.configPath, binding.envName)
		if err != nil {
			return fmt.Errorf("failed to bind env var %s to %s: %w", binding.envName, binding.configPath, err)
		}
	}
	return nil
}
