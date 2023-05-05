package mysql

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"timer/common/conf"
)

var DB *gorm.DB

func GetClient(config conf.MySQLConfig) {
	var err error
	dsn := config.DSN
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	sqlDB, err := DB.DB()
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
}
