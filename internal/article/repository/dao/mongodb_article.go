package dao

import (
	"context"
	"errors"
	"time"

	snowflake "github.com/crazyfrankie/snow-flake"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type MongoArticleDao struct {
	col      *mongo.Collection
	liveCol  *mongo.Collection
	chunkCol *mongo.Collection
	Node     *snowflake.Node
}

func NewMongoArticleDao(db *mongo.Database, node *snowflake.Node) *MongoArticleDao {
	return &MongoArticleDao{
		col:      db.Collection("articles"),
		liveCol:  db.Collection("published_articles"),
		chunkCol: db.Collection("article_chunks"),
		Node:     node,
	}
}

// splitContent 将内容分片
func splitContent(content string) []string {
	if len(content) <= ChunkSize {
		return []string{content}
	}

	var chunks []string
	for len(content) > 0 {
		if len(content) <= ChunkSize {
			chunks = append(chunks, content)
			break
		}
		chunks = append(chunks, content[:ChunkSize])
		content = content[ChunkSize:]
	}
	return chunks
}

func (m *MongoArticleDao) Create(ctx context.Context, art Article) (int64, error) {
	now := time.Now().UnixMilli()
	art.Ctime = now
	art.Utime = now
	id := m.Node.GenerateCode()
	art.ID = uint64(id)

	// 检查内容大小是否需要分片
	contentChunks := splitContent(art.Content)
	if len(contentChunks) > 1 {
		// 需要分片存储
		art.ChunkCount = len(contentChunks)
		art.Content = "" // 清空主文档的内容

		// 开启事务
		session, err := m.col.Database().Client().StartSession()
		if err != nil {
			return 0, err
		}
		defer session.EndSession(ctx)

		_, err = session.WithTransaction(ctx, func(ctx context.Context) (interface{}, error) {
			// 1. 插入主文档
			if _, err := m.col.InsertOne(ctx, art); err != nil {
				return nil, err
			}

			// 2. 插入分片
			var chunks []interface{}
			for i, content := range contentChunks {
				chunk := ArticleChunk{
					ID:        uint64(m.Node.GenerateCode()),
					ArticleID: art.ID,
					Content:   content,
					Order:     i,
					Ctime:     now,
					Utime:     now,
				}
				chunks = append(chunks, chunk)
			}

			_, err := m.chunkCol.InsertMany(ctx, chunks)
			return nil, err
		})

		if err != nil {
			return 0, err
		}
	} else {
		// 不需要分片，直接存储
		art.ChunkCount = 0
		_, err := m.col.InsertOne(ctx, art)
		if err != nil {
			return 0, err
		}
	}

	return int64(id), nil
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

	// 开启事务
	session, err := m.col.Database().Client().StartSession()
	if err != nil {
		return 0, err
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(ctx context.Context) (any, error) {
		// 处理制作库
		if id > 0 {
			err = m.UpdateById(ctx, art)
		} else {
			id, err = m.Create(ctx, art)
			art.ID = uint64(id)
		}
		if err != nil {
			return nil, err
		}

		// 同步到线上库
		now := time.Now().UnixMilli()
		art.Utime = now

		// 如果是分片文章，需要先获取完整内容
		if art.ChunkCount > 0 {
			chunks, err := m.chunkCol.Find(ctx, bson.M{"article_id": art.ID})
			if err != nil {
				return nil, err
			}
			defer chunks.Close(ctx)

			// 组装内容
			var contents = make([]string, art.ChunkCount)
			for chunks.Next(ctx) {
				var chunk ArticleChunk
				if err := chunks.Decode(&chunk); err != nil {
					return nil, err
				}
				contents[chunk.Order] = chunk.Content
			}

			// 组装完整内容
			art.Content = ""
			for _, content := range contents {
				art.Content += content
			}
		}

		// 同步到线上库，使用 upsert 语义
		filter := bson.M{"id": art.ID}
		update := bson.D{
			{Key: "$set", Value: OnlineArticle{
				ID:       art.ID,
				Title:    art.Title,
				Content:  art.Content,
				AuthorID: art.AuthorID,
				Status:   art.Status,
				Ctime:    art.Ctime,
				Utime:    art.Utime,
			}},
			{Key: "$setOnInsert", Value: bson.D{{Key: "ctime", Value: now}}},
		}
		_, err = m.liveCol.UpdateOne(ctx, filter, update)
		return nil, err
	})

	return id, err
}
