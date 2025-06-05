package dao

import (
	"context"
	"errors"
	"fmt"
	"time"

	snowflake "github.com/crazyfrankie/snow-flake"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
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

func (m *MongoArticleDao) Create(ctx context.Context, art MongoArticle) (int64, error) {
	now := time.Now().Unix()
	art.Ctime = now
	art.Utime = now
	id := m.Node.GenerateCode()
	art.ID = uint64(id)

	// 不需要分片，直接存储
	art.ChunkCount = 0
	_, err := m.col.InsertOne(ctx, art)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (m *MongoArticleDao) UpdateById(ctx context.Context, art MongoArticle) error {
	// 操作制作库
	filter := bson.M{"id": art.ID, "author_id": art.AuthorID}
	update := bson.D{
		bson.E{
			Key: "$set",
			Value: bson.M{
				"title":   art.Title,
				"content": art.Content,
				"utime":   time.Now().Unix(),
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

func (m *MongoArticleDao) Sync(ctx context.Context, art MongoArticle) (int64, error) {
	session, err := m.col.Database().Client().StartSession()
	if err != nil {
		return 0, fmt.Errorf("start session failed: %w", err)
	}
	defer session.EndSession(ctx)

	id := int64(art.ID)
	_, err = session.WithTransaction(ctx, func(sessCtx context.Context) (any, error) {
		// 处理制作库
		var err error
		if id > 0 {
			err = m.UpdateById(sessCtx, art)
		} else {
			id, err = m.Create(sessCtx, art)
			art.ID = uint64(id)
		}
		if err != nil {
			return nil, err
		}

		// 分片处理
		now := time.Now().Unix()
		contentChunks := splitContent(art.Content)
		abstract := art.Content
		if len(contentChunks) > 1 {
			abstract = GenerateAbstract(contentChunks[0], 50)
			if err := m.replaceChunks(sessCtx, art.ID, contentChunks, now); err != nil {
				return nil, err
			}
		}

		// 同步线上库
		_, err = m.liveCol.UpdateOne(
			sessCtx,
			bson.M{"id": art.ID},
			bson.D{
				{"$set", bson.M{
					"title":    art.Title,
					"abstract": abstract,
					"chunked":  len(contentChunks) > 1,
					"utime":    now,
				}},
				{"$setOnInsert", bson.M{"ctime": now}},
			},
			options.UpdateOne().SetUpsert(true),
		)
		return nil, err
	})

	return id, err
}

func (m *MongoArticleDao) replaceChunks(ctx context.Context, articleID uint64, chunks []string, now int64) error {
	models := []mongo.WriteModel{
		mongo.NewDeleteManyModel().SetFilter(bson.M{"article_id": articleID}),
	}
	for i, chunk := range chunks {
		models = append(models, mongo.NewInsertOneModel().SetDocument(ArticleChunk{
			ID:        m.Node.GenerateCode(),
			ArticleID: articleID,
			Content:   chunk,
			Order:     i,
			Ctime:     now,
			Utime:     now,
		}))
	}
	_, err := m.chunkCol.BulkWrite(ctx, models)
	return err
}

func (m *MongoArticleDao) GetArticleByID(ctx context.Context, id int64) (MongoArticle, error) {
	filter := bson.M{"id": id}
	cursor, err := m.liveCol.Find(ctx, filter)
	if err != nil {
		return MongoArticle{}, err
	}
	var res MongoArticle
	if err := cursor.All(ctx, &res); err != nil {
		return MongoArticle{}, err
	}
	// 查询经过分片后的结果
	chunks, err := m.chunkCol.Find(ctx, bson.M{"article_id": id})
	if err != nil {
		return MongoArticle{}, err
	}
	defer chunks.Close(ctx)

	// 组装内容
	var contents = make([]string, res.ChunkCount)
	for chunks.Next(ctx) {
		var chunk ArticleChunk
		if err := chunks.Decode(&chunk); err != nil {
			return MongoArticle{}, err
		}
		contents[chunk.Order] = chunk.Content
	}

	res.Content = ""
	for _, content := range contents {
		res.Content += content
	}

	return res, nil
}

func (m *MongoArticleDao) GetArticleList(ctx context.Context, page int, limit int) ([]MongoArticle, error) {
	offset := (page - 1) * limit

	opts := options.Find().
		SetSort(bson.D{{"utime", -1}}).
		SetSkip(int64(offset)).
		SetLimit(int64(limit))
	cursor, err := m.liveCol.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	res := make([]MongoArticle, limit)
	if err := cursor.All(ctx, &res); err != nil {
		return nil, err
	}

	return res, nil
}

// splitContent 将内容分片
func splitContent(content string) []string {
	runes := []rune(content)
	if len(runes) <= ChunkSize {
		return []string{content}
	}

	var chunks []string
	for len(runes) > 0 {
		if len(runes) <= ChunkSize {
			chunks = append(chunks, string(runes))
			break
		}
		// 按字符数切分
		chunks = append(chunks, string(runes[:ChunkSize]))
		runes = runes[ChunkSize:]
	}
	return chunks
}

func GenerateAbstract(text string, maxChars int) string {
	runes := []rune(text)
	if len(runes) <= maxChars {
		return text
	}
	return string(runes[:maxChars]) + "…"
}
