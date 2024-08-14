package dao

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func InitDB() *gorm.DB {
	dsn := "root:root041126@tcp(oj-mysql:11309)/onlinejudge?charset=utf8mb4&parseTime=true&loc=Local"
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
