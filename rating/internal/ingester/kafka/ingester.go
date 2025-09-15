package kafka

import (
	"context"
	"encoding/json"
	"mmoviecom/pkg/logging"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"go.uber.org/zap"

	"mmoviecom/rating/pkg/model"
)

// Ingester defines a Kafka ingester.
type Ingester struct {
	consumer *kafka.Consumer
	topic    string
	logger   *zap.Logger
}

// NewIngester creates a new Kafka ingester.
func NewIngester(addr string, groupID string, topic string, logger *zap.Logger) (*Ingester, error) {
	logger = logger.With(
		zap.String(logging.FieldComponent, "kafka-ingester"),
		zap.String("topic", topic),
	)
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": addr,
		"group.id":          groupID,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return nil, err
	}
	return &Ingester{consumer: consumer, topic: topic, logger: logger}, nil
}

// Ingest starts ingestion from Kafka and returns a channel containing
// rating events representing the data consumed from the topic.
func (i *Ingester) Ingest(ctx context.Context) (chan model.RatingEvent, error) {
	i.logger.Info("Starting Kafka ingester")
	if err := i.consumer.SubscribeTopics([]string{i.topic}, nil); err != nil {
		return nil, err
	}

	ch := make(chan model.RatingEvent, 1)
	go func() {
		for {
			select {
			case <-ctx.Done():
				close(ch)
				i.consumer.Close()
			default:
				msg, err := i.consumer.ReadMessage(-1)
				if err != nil {
					i.logger.Warn("Consumer error", zap.Error(err))
					continue
				}
				i.logger.Info("Processing a message")
				var event model.RatingEvent
				if err := json.Unmarshal(msg.Value, &event); err != nil {
					i.logger.Warn("Unmarshal error", zap.Error(err))
					continue
				}
				ch <- event
			}
		}
	}()
	return ch, nil
}
