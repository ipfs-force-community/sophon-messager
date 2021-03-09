package main

import (
	"io/ioutil"
	"time"

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
		Address: config.AddressConfig{
			RemoteWalletSweepInterval: 10 * time.Second,
		},
		MessageState: config.MessageStateConfig{
			BackTime:          3600 * 24 * time.Second,
			DefaultExpiration: 3600 * 24 * 3 * time.Second,
			CleanupInterval:   3600 * 24 * time.Second,
		},
		MessageService: config.MessageServiceConfig{
			TipsetFilePath: "./tipset.db",
		},
	}
	bytes, _ := toml.Marshal(cfg)
	ioutil.WriteFile("messager.toml", bytes, 0777)
}
