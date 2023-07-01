package config

import (
	"encoding/json"
	"errors"
	"os"
	"path"
)

const (
	defaultConfigDir  = ".binn-cli"
	defaultConfigFile = "config.json"
)

type Config struct {
	Host string `json:"host"`
}

func (c *Config) Save() error {
	if err := validateConfig(c); err != nil {
		return err
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configDirPath := path.Join(homeDir, defaultConfigDir)
	if _, err = os.Stat(configDirPath); os.IsNotExist(err) {
		err := os.Mkdir(configDirPath, 0750)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	f, err := os.OpenFile(path.Join(configDirPath, defaultConfigFile),
		os.O_CREATE|os.O_WRONLY, 0750)
	if err != nil {
		return err
	}

	if err := json.NewEncoder(f).Encode(c); err != nil {
		return err
	}
	return nil
}

func validateConfig(c *Config) error {
	if c.Host == "" {
		return errors.New("host is empty")
	}
	return nil
}

func Load() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	filepath := path.Join(homeDir, defaultConfigDir, defaultConfigFile)
	f, err := os.OpenFile(filepath, os.O_RDONLY, 0750)
	if err != nil {
		return nil, err
	}
	var c Config
	if err := json.NewDecoder(f).Decode(&c); err != nil {
		return nil, err
	}
	return &c, nil
}
