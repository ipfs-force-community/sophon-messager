package models

import (
	"gorm.io/gorm"
)

var _ Repo = (*dbRepo)(nil)

type Repo interface {
	GetDb() *gorm.DB
	DbClose() error
}

type dbRepo struct {
	*gorm.DB
}

func (d dbRepo) GetDb() *gorm.DB {
	return d.DB
}

func (d dbRepo) DbClose() error {
	return d.DbClose()
}
