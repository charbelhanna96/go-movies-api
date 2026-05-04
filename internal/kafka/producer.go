// Package kafka provides Kafka producer functionality.
package kafka

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/IBM/sarama"
	commonkafka "github.com/charbelhanna96/go-movies-common/pkg/kafka"
)

// KafkaProducer defines the interface for publishing Kafka events.
type KafkaProducer interface {
	PublishSearchEvent(event commonkafka.SearchEvent) error
}

// Producer wraps a sarama SyncProducer.
type Producer struct {
	producer sarama.SyncProducer
}

// NewProducer creates a new Kafka producer.
func NewProducer(brokers []string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 3

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("create kafka producer: %w", err)
	}

	slog.Info("kafka producer connected", "brokers", brokers)

	return &Producer{producer: producer}, nil
}

// PublishSearchEvent publishes a movie search event to Kafka.
func (p *Producer) PublishSearchEvent(event commonkafka.SearchEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal search event: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: commonkafka.TopicMovieSearched,
		Value: sarama.ByteEncoder(payload),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("publish search event: %w", err)
	}

	slog.Debug("search event published",
		"topic", commonkafka.TopicMovieSearched,
		"partition", partition,
		"offset", offset,
	)

	return nil
}

// Close shuts down the producer.
func (p *Producer) Close() error {
	return p.producer.Close()
}
