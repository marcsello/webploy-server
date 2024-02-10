package config

import (
	"github.com/creasty/defaults"
	"gitlab.com/MikeTTh/env"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"os"
)

const ConfigDefaultPath = "/etc/webploy/webploy.conf"

func LoadConfig(logger *zap.Logger) (WebployConfig, error) {
	var err error

	// open file
	configPath := env.String("WEBPLOY_CONFIG", ConfigDefaultPath)
	logger.Info("Loading config", zap.String("configPath", configPath))

	var configFile *os.File
	configFile, err = os.Open(configPath) // #nosec G304
	if err != nil {
		return WebployConfig{}, err
	}
	defer func(configFile *os.File) {
		e := configFile.Close()
		if e != nil {
			logger.Warn("could not close log file", zap.Error(e))
		}
	}(configFile)

	// parse it
	var newConfig WebployConfig

	// read defaults
	err = defaults.Set(&newConfig)
	if err != nil {
		return WebployConfig{}, err
	}

	// decode file
	decoder := yaml.NewDecoder(configFile)
	decoder.KnownFields(true)
	err = decoder.Decode(&newConfig)
	if err != nil {
		return WebployConfig{}, err
	}

	logger.Debug("Config successfully loaded", zap.Any("config", newConfig))
	// very good
	return newConfig, nil
}
