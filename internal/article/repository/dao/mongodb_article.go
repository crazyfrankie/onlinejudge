package dao

import (
	"context"
	"errors"

	"time"

	"github.com/crazyfrankie/snow-flake"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type MongoArticleDao struct {
	col     *mongo.Collection
	liveCol *mongo.Collection
	Node    *snowflake.Node
}

func NewMongoArticleDao(db *mongo.Database, node *snowflake.Node) *MongoArticleDao {
	return &MongoArticleDao{
		col:     db.Collection("articles"),
		liveCol: db.Collection("published_articles"),
		Node:    node,
	}
}

func (m *MongoArticleDao) Create(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	id := m.Node.GenerateCode()
	art.ID = uint64(id)

	_, err := m.col.InsertOne(ctx, art)

	return id, err
}

func (m *MongoArticleDao) UpdateById(ctx context.Context, art Article) error {
	// 操作制作库
	filter := bson.M{"id": art.ID, "author_id": art.AuthorID}
	update := bson.D{
		bson.E{
			Key: "$set",
			Value: bson.M{
				"title":   art.Title,
				"content": art.Content,
				"utime":   time.Now().UnixMilli(),
				"status":  art.Status,
			}},
	}

	res, err := m.col.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if res.ModifiedCount == 0 {
		return errors.New("更新数据失败")
	}

	return nil
}

func (m *MongoArticleDao) Sync(ctx context.Context, art Article) (int64, error) {
	var (
		id  = int64(art.ID)
		err error
	)
	if id > 0 {
		err := m.UpdateById(ctx, art)
		if err != nil {
			return 0, err
		}
	} else {
		id, err = m.Create(ctx, art)
	}
	if err != nil {
		return 0, err
	}
	art.ID = uint64(id)
	// 操作线上库, upsert 语义
	now := time.Now().UnixMilli()
	art.Utime = now
	//update := bson.M{
	//	"$set":         art,
	//	"$setOnInsert": bson.D{bson.E{Key: "ctime", Value: now}},
	//}
	update := bson.E{Key: "$set", Value: OnlineArticle(art)}
	upsert := bson.E{Key: "$setOnInsert", Value: bson.D{bson.E{Key: "ctime", Value: now}}}
	filter := bson.M{"id": art.ID}
	_, err = m.liveCol.UpdateOne(ctx, filter, bson.D{update, upsert})

	return id, err
}

//func (m *MongoArticleDao) SyncStatus(ctx context.Context, author, id uint64, status domain.ArticleStatus) error {
//	now := time.Now().UnixMilli()
//
//	filter := bson.M{"id": id, "author_id": author}
//	update := bson.D{
//		bson.E{
//			Key: "$set",
//			Value: bson.M{
//				"utime":  now,
//				"status": status.ToUint8(),
//			},
//		},
//	}
//	_, err := m.col.UpdateOne(ctx, filter, update)
//	if err != nil {
//		return nil
//	}
//
//	_, err = m.liveCol.UpdateOne(ctx, filter, update)
//
//	return err
//}
