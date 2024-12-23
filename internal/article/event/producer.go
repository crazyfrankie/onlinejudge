package event

import (
	"context"
	"encoding/json"

	"github.com/IBM/sarama"
)

type ArticleProducer struct {
	SyncProducer sarama.SyncProducer
}

func NewArticleProducer(SyncProducer sarama.SyncProducer) Producer {
	return &ArticleProducer{
		SyncProducer: SyncProducer,
	}
}

// ProduceReadEvent 如果说重试逻辑很复杂，使用装饰器
// 如果逻辑很简单，直接在这里写
func (a *ArticleProducer) ProduceReadEvent(ctx context.Context, evt ReadEvent) error {
	data, err := json.Marshal(evt)
	if err != nil {
		return err
	}

	_, _, err = a.SyncProducer.SendMessage(&sarama.ProducerMessage{
		Topic: "article_read",
		Value: sarama.ByteEncoder(data),
	})

	return err
}
