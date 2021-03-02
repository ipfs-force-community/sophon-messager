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
