package ioc

import (
	"context"
	snowflake "github.com/crazyfrankie/snow-flake"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
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

	return nil
}
