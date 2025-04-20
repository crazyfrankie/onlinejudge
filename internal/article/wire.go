package article

import (
	"github.com/IBM/sarama"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/crazyfrankie/onlinejudge/internal/article/event"
	"github.com/crazyfrankie/onlinejudge/internal/article/repository"
	"github.com/crazyfrankie/onlinejudge/internal/article/repository/cache"
	"github.com/crazyfrankie/onlinejudge/internal/article/repository/dao"
	"github.com/crazyfrankie/onlinejudge/internal/article/service"
	"github.com/crazyfrankie/onlinejudge/internal/article/web"
	"github.com/crazyfrankie/onlinejudge/pkg/zapx"
)

func NewSyncProducer(client sarama.Client) sarama.SyncProducer {
	res, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		panic(err)
	}

	return res
}

func InitModule(db *gorm.DB, cmd redis.Cmdable, client sarama.Client, l *zapx.Logger) *Module {
	wire.Build(
		dao.NewArticleDao,
		dao.NewInteractiveDao,
		cache.NewInteractiveCache,

		repository.NewArticleRepository,
		repository.NewInteractiveArtRepository,

		NewSyncProducer,
		event.NewArticleProducer,
		event.NewArticleConsumer,

		service.NewArticleService,
		service.NewInteractiveService,

		web.NewArticleHandler,
		web.NewAdminHandler,

		wire.Struct(new(Module), "*"),
	)
	return new(Module)
}
