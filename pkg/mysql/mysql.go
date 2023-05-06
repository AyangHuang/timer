package mysql

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"timer/common/conf"
)

var db *gorm.DB

func GetClient(config *conf.MySQLConfig) *gorm.DB {
	var err error
	dsn := config.DSN
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	sqlDB, err := db.DB()
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	return db
}
