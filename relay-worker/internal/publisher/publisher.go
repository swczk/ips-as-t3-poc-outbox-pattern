package publisher

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type Publisher struct {
	writer *kafka.Writer
}

func New(brokers, topic string) *Publisher {
	return &Publisher{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers),
			Topic:        topic,
			Balancer:     &kafka.Hash{},
			RequiredAcks: kafka.RequireOne,
		},
	}
}

func (p *Publisher) Publish(ctx context.Context, key, value []byte) error {
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: value,
	})
}

func (p *Publisher) Close() error {
	return p.writer.Close()
}
