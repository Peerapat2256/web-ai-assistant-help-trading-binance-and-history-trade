package query

import "gorm.io/gorm"

var defaultDB *gorm.DB

func SetDefault(db *gorm.DB) {
	defaultDB = db
}

func DB() *gorm.DB {
	return defaultDB
}
