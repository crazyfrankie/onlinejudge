package event

import (
	"context"
	"github.com/IBM/sarama"
	"go.uber.org/zap"
	"oj/internal/article/repository"
	"oj/pkg/saramax"
	"time"
)

type ArticleConsumer struct {
	client sarama.Client
	repo   *repository.InteractiveArtRepository
	l      *zap.Logger
}

func NewArticleConsumer(client sarama.Client, repo *repository.InteractiveArtRepository, l *zap.Logger) *ArticleConsumer {
	return &ArticleConsumer{
		client: client,
		repo:   repo,
		l:      l,
	}
}

func (a *ArticleConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("interactive_article", a.client)
	if err != nil {
		return err
	}

	go func() {
		err := cg.Consume(context.Background(), []string{"article_read"}, saramax.NewHandler[ReadEvent](a.l, a.Consume))
		if err != nil {
			a.l.Error("退出消费循环异常", zap.Error(err))
		}
	}()

	return err
}

func (a *ArticleConsumer) Consume(msg *sarama.ConsumerMessage, t ReadEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	return a.repo.IncrReadCnt(ctx, "article", t.Aid)
}
