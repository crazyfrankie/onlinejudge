package ioc

import (
	"fmt"
	"os"
	"time"

	prometheus2 "github.com/prometheus/client_golang/prometheus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"gorm.io/plugin/prometheus"

	"github.com/crazyfrankie/onlinejudge/config"
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

	// prometheus 埋点
	err = db.Use(prometheus.New(prometheus.Config{
		DBName: "onlinejudge",
		// Prometheus 获取数据的间隔
		RefreshInterval: 15,
		MetricsCollector: []prometheus.MetricsCollector{
			&prometheus.MySQL{},
		},
	}))
	if err != nil {
		panic(err)
	}

	// 监控 GORM 的执行时间
	callbacks := newCallbacks()
	callbacks.registerAll(db)

	return db
}

type Callbacks struct {
	vector *prometheus2.SummaryVec
}

func newCallbacks() *Callbacks {
	vector := prometheus2.NewSummaryVec(prometheus2.SummaryOpts{
		Namespace: "cfcstudio_frank",
		Subsystem: "onlinejudge",
		Name:      "gorm_query_time",
		Help:      "统计 GORM 的执行时间",
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.9:   0.01,
			0.99:  0.001,
			0.999: 0.0001,
		},
	}, []string{"type", "table"})
	prometheus2.MustRegister(vector)

	return &Callbacks{
		vector: vector,
	}
}

func (c *Callbacks) before() func(db *gorm.DB) {
	return func(db *gorm.DB) {
		// 记录时间
		startTime := time.Now()
		db.Set("start_time", startTime)
	}
}

func (c *Callbacks) after(typ string) func(db *gorm.DB) {
	return func(db *gorm.DB) {
		val, _ := db.Get("start_time")
		startTime, ok := val.(time.Time)
		if !ok {
			// 啥也干不了
			// 顶多打日志
			return
		}
		duration := time.Since(startTime)
		// 上报 prometheus
		table := db.Statement.Table
		if table == "" {
			table = "unknown"
		}
		c.vector.WithLabelValues(typ, table).Observe(float64(duration.Milliseconds()))
	}
}

func (c *Callbacks) registerAll(db *gorm.DB) {
	// 钩子函数

	// 增
	err := db.Callback().Create().Before("*").Register("prometheus_create_before", c.before())
	if err != nil {
		panic(err)
	}

	err = db.Callback().Create().After("*").Register("prometheus_create_after", c.after("create"))
	if err != nil {
		panic(err)
	}

	// 改
	err = db.Callback().Update().Before("*").Register("prometheus_update_before", c.before())
	if err != nil {
		panic(err)
	}

	err = db.Callback().Update().After("*").Register("prometheus_update_after", c.after("update"))
	if err != nil {
		panic(err)
	}

	// 删
	err = db.Callback().Delete().Before("*").Register("prometheus_delete_before", c.before())
	if err != nil {
		panic(err)
	}

	err = db.Callback().Delete().After("*").Register("prometheus_delete_after", c.after("delete"))
	if err != nil {
		panic(err)
	}

	// 查
	err = db.Callback().Query().Before("*").Register("prometheus_query_before", c.before())
	if err != nil {
		panic(err)
	}

	err = db.Callback().Query().After("*").Register("prometheus_query_after", c.after("delete"))
	if err != nil {
		panic(err)
	}

	// 原生 SQL
	err = db.Callback().Raw().Before("*").Register("prometheus_raw_before", c.before())
	if err != nil {
		panic(err)
	}

	err = db.Callback().Raw().After("*").Register("prometheus_raw_after", c.after("raw"))
	if err != nil {
		panic(err)
	}

	// 返回单条记录
	err = db.Callback().Row().Before("*").Register("prometheus_row_before", c.before())
	if err != nil {
		panic(err)
	}

	err = db.Callback().Row().After("*").Register("prometheus_row_after", c.after("row"))
	if err != nil {
		panic(err)
	}
}
