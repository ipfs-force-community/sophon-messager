package repo

import (
	"gorm.io/gorm"
)

type Repo interface {
	GetDb() *gorm.DB
	DbClose() error
	AutoMigrate() error

	WalletRepo() WalletRepo
	MessageRepo() MessageRepo
}
