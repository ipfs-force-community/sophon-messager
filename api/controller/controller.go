package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/ipfs-force-community/venus-messager/models/repo"
	"github.com/sirupsen/logrus"
)

type BaseController struct {
	Context *gin.Context
	Repo    repo.Repo
	Logger  *logrus.Logger
}
