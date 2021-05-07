package main

import (
	"io/ioutil"

	"github.com/pelletier/go-toml"

	"github.com/filecoin-project/venus-messager/config"
)

func main() {
	bytes, _ := toml.Marshal(config.DefaultConfig())
	_ = ioutil.WriteFile("messager.toml", bytes, 0777)
}
