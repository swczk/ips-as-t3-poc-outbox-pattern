package publisher

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type Publisher struct {
	writer   *kafka.Writer
	dlWriter *kafka.Writer
}

func New(brokers, deadLetterTopic string) *Publisher {
	addr := kafka.TCP(brokers)
	return &Publisher{
		writer: &kafka.Writer{
			Addr:         addr,
			Balancer:     &kafka.Hash{},
			RequiredAcks: kafka.RequireOne,
		},
		dlWriter: &kafka.Writer{
			Addr:         addr,
			Topic:        deadLetterTopic,
			Balancer:     &kafka.Hash{},
			RequiredAcks: kafka.RequireOne,
		},
	}
}

func (p *Publisher) Publish(ctx context.Context, topic string, key, value []byte) error {
	return p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   key,
		Value: value,
	})
}

func (p *Publisher) PublishDeadLetter(ctx context.Context, key, value []byte) error {
	return p.dlWriter.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: value,
	})
}

func (p *Publisher) Close() {
	_ = p.writer.Close()
	_ = p.dlWriter.Close()
}
