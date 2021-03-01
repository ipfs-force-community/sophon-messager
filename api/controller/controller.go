package controller

import (
	"github.com/ipfs-force-community/venus-messager/models/repo"
	"github.com/sirupsen/logrus"
)

type BaseController struct {
	Repo   repo.Repo
	Logger *logrus.Logger
}
