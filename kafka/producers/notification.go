package kafka

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

type KafkaNotificationProducer struct {
	Writer *kafka.Writer
}

func NewKafkaNotificationProducer(brokers []string, topic string) *KafkaNotificationProducer {

	return &KafkaNotificationProducer{
		Writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
			Async: true,
		},
	}
}

func (p *KafkaNotificationProducer) Send(ctx context.Context, data []byte) error {
	err := p.Writer.WriteMessages(ctx, kafka.Message{Value: data})
	if err != nil {
		log.Printf(" Kafka Write Error: %v\n", err)
	}
	return err
}

func (p *KafkaNotificationProducer) Close() error {
	return p.Writer.Close()
}
