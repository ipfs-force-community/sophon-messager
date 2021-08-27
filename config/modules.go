package config

import (
	"io/ioutil"
	"os"

	"github.com/pelletier/go-toml"
)

func ConfigExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if !os.IsExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
func ReadConfig(path string) (*Config, error) {
	configBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	config := new(Config)
	if err := toml.Unmarshal(configBytes, config); err != nil {
		return nil, err
	}
	dur := config.MessageService.WaitingChainHeadStableDuration

	if dur == 0 {
		dur = DefWaitingChainHeadStableDuration
	} else if dur < MinWaitingChainHeadStableDuration {
		dur = MinWaitingChainHeadStableDuration
	} else if dur > MaxWaitingChainHeadStableDuration {
		dur = MaxWaitingChainHeadStableDuration
	}

	config.MessageService.WaitingChainHeadStableDuration = dur

	return config, nil
}

func WriteConfig(path string, cfg *Config) error {
	cfgBytes, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, cfgBytes, 0666)
}

func CheckFile(cfg *Config) error {
	if _, err := os.Stat(cfg.MessageService.TipsetFilePath); err != nil {
		if os.IsNotExist(err) {
			_, err := os.Create(cfg.MessageService.TipsetFilePath)
			return err
		}
		return err
	}

	return nil
}
