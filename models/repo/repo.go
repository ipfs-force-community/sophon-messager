package repo

import (
	"gorm.io/gorm"
)

type Repo interface {
	GetDb() *gorm.DB
	DbClose() error
	MessageRepo() MessageRepo
	AutoMigrate() error
}
