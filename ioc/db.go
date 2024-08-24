package ioc

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"oj/config"
)

func InitDB() *gorm.DB {
	dsn := config.Config.DB.DSN
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 表名不加s
		},
	})
	if err != nil {
		panic("failed to connect database")
	}
	return db
}
