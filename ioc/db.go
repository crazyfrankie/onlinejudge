package ioc

import (
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"oj/config"
	"oj/internal/problem/repository/dao"
)

func InitDB() *gorm.DB {
	dsn := config.Config.DB.DSN
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 表名不加s
		},
		// 可设置外键约束
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		panic("failed to connect database")
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic("failed to connect database")
	}
	// 设置连接池参数
	sqlDB.SetMaxIdleConns(config.Config.DB.MaxIdleConns)                                    // 最大空闲连接数
	sqlDB.SetMaxOpenConns(config.Config.DB.MaxOpenConns)                                    // 最大打开连接数
	sqlDB.SetConnMaxLifetime(time.Duration(config.Config.DB.ConnMaxLifetime) * time.Minute) // 连接的最大生命周期
	Migrate(db)

	return db
}

func Migrate(db *gorm.DB) error {
	// 自动迁移，创建表
	if err := db.AutoMigrate(&dao.Problem{}); err != nil {
		return err
	}
	return nil
}
