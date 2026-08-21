package dao

import (
	"phone_ios_backend/connection"
	"gorm.io/gorm"
)

type DBConnection struct {
	db *gorm.DB
}

func NewDB() *DBConnection {
	dc := &DBConnection{}
	dc.db = connection.DbConnection
	return dc
}
