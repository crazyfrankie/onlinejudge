package ioc

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	snowflake "github.com/crazyfrankie/snow-flake"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func InitMongoDB() (*mongo.Database, *snowflake.Node) {
	monitor := &event.CommandMonitor{}

	opt := options.Client().ApplyURI("mongodb://localhost:27017").
		SetMonitor(monitor)
	client, err := mongo.Connect(opt)
	if err != nil {
		panic(err)
	}

	db := client.Database("onlinejudge")
	err = InitCollections(db)
	if err != nil {
		panic(err)
	}

	node, err := snowflake.NewNode(1)
	if err != nil {
		panic(err)
	}

	return db, node
}

func InitCollections(db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	err := db.CreateCollection(ctx, "articles")
	if err != nil {
		return err
	}

	err = db.CreateCollection(ctx, "published_articles")
	if err != nil {
		return err
	}

	err = db.CreateCollection(ctx, "article_chunks")
	if err != nil {
		return err
	}

	// 创建分片相关索引
	_, err = db.Collection("article_chunks").Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "article_id", Value: 1}, {Key: "order", Value: 1}},
		},
	})

	return err
}
