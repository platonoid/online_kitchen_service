package kafka

import (
	"context"
	"encoding/json"
	"log"

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
