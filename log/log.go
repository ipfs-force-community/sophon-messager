package log

import (
	"context"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"

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
		return nil, fmt.Errorf("open log file fail %v", err)
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
