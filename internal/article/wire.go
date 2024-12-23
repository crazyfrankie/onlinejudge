//go:build wireinject

package article

import (
	"github.com/IBM/sarama"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"oj/internal/article/event"

	"oj/internal/article/repository"
	"oj/internal/article/repository/cache"
	"oj/internal/article/repository/dao"
	"oj/internal/article/service"
	"oj/internal/article/web"
)

func NewSyncProducer(client sarama.Client) sarama.SyncProducer {
	res, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		panic(err)
	}

	return res
}

func InitConsumer(db *gorm.DB, cmd redis.Cmdable, client sarama.Client, l *zap.Logger) *event.ArticleConsumer {
	wire.Build(
		dao.NewInteractiveDao,
		cache.NewInteractiveCache,

		repository.NewInteractiveArtRepository,

		event.NewArticleConsumer,
	)
	return new(event.ArticleConsumer)
}

func InitArticleHandler(db *gorm.DB, cmd redis.Cmdable, client sarama.Client, l *zap.Logger) *web.ArticleHandler {
	wire.Build(
		dao.NewArticleDao,
		dao.NewInteractiveDao,
		cache.NewInteractiveCache,

		repository.NewArticleRepository,
		repository.NewInteractiveArtRepository,

		NewSyncProducer,

		event.NewArticleProducer,
		service.NewArticleService,
		service.NewInteractiveService,

		web.NewArticleHandler,
	)
	return new(web.ArticleHandler)
}
