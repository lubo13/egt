package processor

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type MessageProcessorInterface interface {
	Process(ctx context.Context, kafkaMessage *kafka.Message) error
}
