package main

import (
	"io/ioutil"

	"github.com/pelletier/go-toml"

	"github.com/ipfs-force-community/venus-messager/config"
)

func main() {
	bytes, _ := toml.Marshal(config.DefaultConfig())
	_ = ioutil.WriteFile("messager.toml", bytes, 0777)
}
