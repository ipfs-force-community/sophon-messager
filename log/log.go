package log

import (
	"context"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
	"os"

	"github.com/filecoin-project/venus-messager/config"
)

type Logger struct {
	*logrus.Logger
}

func New() *Logger {
	return &Logger{logrus.New()}
}

func SetLogger(logCfg *config.LogConfig) (*Logger, error) {
	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})
	logger := &Logger{log}
	err := logger.SetLogLevel(context.Background(), logCfg.Level)
	if err != nil {
		return nil, err
	}

	file, err := os.OpenFile(logCfg.Path, os.O_CREATE|os.O_WRONLY, 0666)
	if err == nil {
		logrus.SetOutput(file)
	} else {
		return nil, xerrors.Errorf("open log file fail %v", err)
	}
	return logger, nil
}

func (logger *Logger) SetLogLevel(ctx context.Context, levelStr string) error {
	level, err := logrus.ParseLevel(levelStr)
	if err != nil {
		return err
	}
	logger.SetLevel(level)

	return nil
}

// 2016-09-27 09:38:21.541541811 +0200 CEST
// 127.0.0.1 - frank [2021-04-09 15:58:00]
// "GET /apache_pb.gif HTTP/1.0" 200 2326
// "http://www.example.com/start.html"
// "Mozilla/4.08 [en] (Win98; I ;Nav)"
// copy from https://github.com/toorop/gin-logrus/blob/master/logger.go

var timeFormat = "2006-01-02 15:04:05"
