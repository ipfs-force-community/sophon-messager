package main

import (
	"github.com/ipfs-force-community/venus-messager/config"
	"github.com/pelletier/go-toml"
	"io/ioutil"
)

func main() {
	cfg := config.Config{
		DbConfig: config.DbConfig{
			Type:   "sqlite",
			MySql:  config.MySqlConfig{},
			Sqlite: config.SqliteConfig{Path: "./message.db"},
		},
		JWT: config.JWTConfig{},
		Log: config.LogConfig{
			Path: "messager.log",
		},
		API: config.APIConfig{
			Address: "127.0.0.1:39812",
		},
	}
	bytes, _ := toml.Marshal(cfg)
	ioutil.WriteFile("example.toml", bytes, 0777)
}
