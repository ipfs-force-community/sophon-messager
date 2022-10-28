package utils

import (
	"os"

	logging "github.com/ipfs/go-log/v2"
)

func SetupLogLevels() {
	val, set := os.LookupEnv("GOLOG_LOG_LEVEL")
	if !set {
		_ = logging.SetLogLevel("*", "INFO")
	} else {
		_ = logging.SetLogLevel("*", val)
	}
}
