package main

import (
	"io/ioutil"

	"github.com/pelletier/go-toml"

	"github.com/ipfs-force-community/venus-messager/config"
)

func main() {
	cfg := config.Config{
		DB: config.DbConfig{
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
		Node: config.NodeConfig{
			Url:   "",
			Token: "",
		},
	}
	bytes, _ := toml.Marshal(cfg)
	ioutil.WriteFile("example.toml", bytes, 0777)
}
