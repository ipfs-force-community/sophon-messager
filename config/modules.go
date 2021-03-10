package config

import (
	"io/ioutil"
	"os"

	"github.com/pelletier/go-toml"
)

func ReadConfig(path string) (*Config, error) {
	configBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	config := new(Config)
	if err := toml.Unmarshal(configBytes, config); err != nil {
		return nil, err
	}
	return config, nil
}

func CheckFile(cfg *Config) error {
	if _, err := os.Stat(cfg.MessageService.TipsetFilePath); err != nil {
		if os.IsNotExist(err) {
			if _, err := os.Create(cfg.MessageService.TipsetFilePath); err != nil {
				return err
			}
		}
		return err
	}

	return nil
}
