package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/ipfs-force-community/venus-messager/models"
	"github.com/sirupsen/logrus"
)

type BaseController struct {
	Context *gin.Context
	Repo    models.Repo
	Logger  *logrus.Logger
}
