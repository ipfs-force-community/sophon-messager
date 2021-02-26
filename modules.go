package main

import (
	"github.com/ipfs-force-community/venus-messager/config"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
	"os"
)

type ShutdownChan chan struct{}

func SetLogger(logCfg *config.LogConfig) (*logrus.Logger, error) {
	log := logrus.New()
	file, err := os.OpenFile(logCfg.Path, os.O_CREATE|os.O_WRONLY, 0666)
	if err == nil {
		logrus.SetOutput(file)
	} else {
		return nil, xerrors.Errorf("open log file fail")
	}
	return log, nil
}
