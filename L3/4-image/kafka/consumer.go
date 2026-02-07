package kafka

import (
	"bytes"
	"context"
	"encoding/gob"
	"time"
	"wildberries-go-course/L3-4/database"
	"wildberries-go-course/L3-4/model"
	"wildberries-go-course/L3-4/repository"
	"wildberries-go-course/L3-4/storage"
	"wildberries-go-course/L3-4/util"

	kafka_lib "github.com/segmentio/kafka-go"
	"github.com/wb-go/wbf/kafka"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"
)

type ImageConsumer struct {
	consumer      *kafka.Consumer
	imageStorage  *storage.ImageStorage
	retryStrategy retry.Strategy
}

func NewImageConsumer(brokers []string, topic, groupID string, imageStorage *storage.ImageStorage) *ImageConsumer {
	strategy := retry.Strategy{
		Attempts: 3,
		Delay:    time.Millisecond,
		Backoff:  5,
	}
	return &ImageConsumer{
		consumer:      kafka.NewConsumer(brokers, topic, groupID),
		imageStorage:  imageStorage,
		retryStrategy: strategy,
	}
}

func (c *ImageConsumer) Start(ctx context.Context, db *database.Database) error {
	defer c.consumer.Close()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	msgChan := make(chan kafka_lib.Message)

	c.consumer.StartConsuming(ctx, msgChan, c.retryStrategy)

	zlog.Logger.Info().Msg("Image consumer started")

	for msg := range msgChan {
		zlog.Logger.Info().Msg("Received image message from kafka")
		id := string(msg.Key)
		if err := c.processImageMessage(msg); err != nil {
			zlog.Logger.Error().Msgf("Fialed to process message with key %s: %s", id, err.Error())
			repository.UpdateImageProcess(ctx, db, id, model.Failed)
			if err := c.consumer.Commit(ctx, msg); err != nil {
				zlog.Logger.Error().Msgf("Failed to commit failed message: %s", err.Error())
			}
			continue
		}
		if err := repository.UpdateImageProcess(ctx, db, id, model.Processed); err != nil {
			zlog.Logger.Error().Msgf("Fialed to update status for image with key %s: %s", id, err.Error())
			continue
		}
		if err := c.consumer.Commit(ctx, msg); err != nil {
			zlog.Logger.Error().Msgf("Failed to commit: %s", err.Error())
		}
	}
	zlog.Logger.Info().Msg("Consumer stopped")
	return nil
}

func (c *ImageConsumer) processImageMessage(msg kafka_lib.Message) error {
	var imageInfo model.ImageInfo
	buf := bytes.NewBuffer(msg.Value)
	decoder := gob.NewDecoder(buf)

	if err := decoder.Decode(&imageInfo); err != nil {
		return err
	}
	imageData, err := c.imageStorage.GetImageData(&imageInfo)
	if err != nil {
		return err
	}
	operatedImageData, err := util.OperateOnImage(imageInfo, imageData)

	imageInfo.Type = model.Modified

	if err != nil {
		return err
	}

	c.imageStorage.StoreImage(imageInfo, operatedImageData)
	return nil
}
