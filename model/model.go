package model

import "gorm.io/gorm"

var db *gorm.DB

func Use() *gorm.DB {
	return db
}

func InitDB(d *gorm.DB) {
	db = d
}
