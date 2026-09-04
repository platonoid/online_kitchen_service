package kafka

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"strings"

	"github.com/segmentio/kafka-go"
	"kitchen/internal/domain"
)

type LoggingPublisher struct {
	Logger *log.Logger
}

func NewLoggingPublisher(logger *log.Logger) *LoggingPublisher {
	if logger == nil {
		logger = log.Default()
	}
	return &LoggingPublisher{Logger: logger}
}

func (p *LoggingPublisher) Publish(ctx context.Context, event domain.OrderEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	p.Logger.Printf("kafka event %s: %s", event.Type, data)
	return nil
}

type LoggingEventPublisher = LoggingPublisher

func NewLoggingEventPublisher(logger *log.Logger) *LoggingPublisher {
	return NewLoggingPublisher(logger)
}

type KafkaPublisher struct {
	writer *kafka.Writer
}

func NewKafkaPublisher(brokers, topic string) *KafkaPublisher {
	return &KafkaPublisher{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(strings.Split(brokers, ",")...),
			Topic:    topic,
			Balancer: &kafka.Hash{},
		},
	}
}

func (p *KafkaPublisher) Publish(ctx context.Context, event domain.OrderEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(strconv.Itoa(event.OrderID)),
		Value: data,
	})
}

func (p *KafkaPublisher) Close() error {
	return p.writer.Close()
}

type Consumer struct {
	reader *kafka.Reader
}

func NewConsumer(brokers, topic, groupID string) *Consumer {
	return &Consumer{reader: kafka.NewReader(kafka.ReaderConfig{
		Brokers: strings.Split(brokers, ","),
		Topic:   topic,
		GroupID: groupID,
	})}
}

func (c *Consumer) Run(ctx context.Context, handler func(context.Context, domain.OrderEvent) error) error {
	for {
		message, err := c.reader.ReadMessage(ctx)
		if err != nil {
			return err
		}
		var event domain.OrderEvent
		if err := json.Unmarshal(message.Value, &event); err != nil {
			return err
		}
		if err := handler(ctx, event); err != nil {
			return err
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
