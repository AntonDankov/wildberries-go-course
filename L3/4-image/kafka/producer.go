package kafka

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"net"
	"strconv"
	"time"
	"wildberries-go-course/L3-4/model"

	kafkago "github.com/segmentio/kafka-go"
	"github.com/wb-go/wbf/kafka"
	"github.com/wb-go/wbf/retry"
)

type ImageBrokerProducer struct {
	producer *kafka.Producer
}

func NewImageProducer(brokers []string, topic string) *ImageBrokerProducer {
	return &ImageBrokerProducer{
		producer: kafka.NewProducer(brokers, topic),
	}
}

func (p *ImageBrokerProducer) SendImage(ctx context.Context, image *model.ImageInfo) error {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)

	if err := encoder.Encode(image); err != nil {
		return err
	}

	strategy := retry.Strategy{
		Attempts: 3,
		Delay:    time.Millisecond,
		Backoff:  5,
	}

	return p.producer.SendWithRetry(ctx, strategy, []byte(image.ID), buf.Bytes())
}

func (p *ImageBrokerProducer) Close() error {
	return p.producer.Close()
}

func CreateTopic(brokers []string, topic string, numPartitions int, replicationFactor int) error {
	conn, err := kafkago.Dial("tcp", brokers[0])
	if err != nil {
		return fmt.Errorf("failed to dial kafka: %w", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("failed to get controller: %w", err)
	}

	controllerConn, err := kafkago.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return fmt.Errorf("failed to dial controller: %w", err)
	}
	defer controllerConn.Close()

	topicConfigs := []kafkago.TopicConfig{
		{
			Topic:             topic,
			NumPartitions:     numPartitions,
			ReplicationFactor: replicationFactor,
		},
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		return fmt.Errorf("failed to create topic: %w", err)
	}

	return nil
}
