package postgres

import (
	"gorm.io/gorm"
)

type Db struct {
	*gorm.DB
}

var dbInstance *Db

func GetCluster() *Db {
	return dbInstance
}

func SetCluster(cluster *gorm.DB) {
	dbInstance = &Db{cluster}
}
