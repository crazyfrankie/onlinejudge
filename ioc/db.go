package ioc

import (
	"fmt"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/crazyfrankie/onlinejudge/config"
	"github.com/crazyfrankie/onlinejudge/internal/article/repository/dao"
)

func InitDB() *gorm.DB {
	dsn := fmt.Sprintf(config.GetConf().MySQL.DSN,
		os.Getenv("MYSQL_USER"),
		os.Getenv("MYSQL_PASSWORD"),
		os.Getenv("MYSQL_HOST"))
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
	sqlDB.SetMaxIdleConns(config.GetConf().MySQL.MaxIdleConns)                                    // 最大空闲连接数
	sqlDB.SetMaxOpenConns(config.GetConf().MySQL.MaxOpenConns)                                    // 最大打开连接数
	sqlDB.SetConnMaxLifetime(time.Duration(config.GetConf().MySQL.ConnMaxLifeTime) * time.Minute) // 连接的最大生命周期
	err = Migrate(db)
	if err != nil {
		panic(err)
	}

	return db
}

func Migrate(db *gorm.DB) error {
	// 自动迁移，创建表
	if err := db.AutoMigrate(&dao.OnlineArticle{}); err != nil {
		return err
	}
	return nil
}
