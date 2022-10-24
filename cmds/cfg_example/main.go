package main

import (
	"os"

	"github.com/pelletier/go-toml"

	"github.com/filecoin-project/venus-messager/config"
)

func main() {
	bytes, _ := toml.Marshal(config.DefaultConfig())
	_ = os.WriteFile("messager.toml", bytes, 0o777)
}
