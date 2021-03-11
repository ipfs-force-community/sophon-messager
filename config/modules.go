package config

import (
	"io/ioutil"

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

func WriteFile(path string, cfg Config) error {
	b, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, b, 0666)
}
