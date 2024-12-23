package event

import "context"

type Producer interface {
	ProduceReadEvent(ctx context.Context, evt ReadEvent) error
}

type ReadEvent struct {
	Uid uint64
	Aid uint64
}

type Consumer interface {
	Start() error
}
