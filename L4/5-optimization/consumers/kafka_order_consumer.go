package consumers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"wildberries-go-course/L0/model"
	"wildberries-go-course/L0/repository"

	"github.com/IBM/sarama"
)

func RunDeliveryConsumer(repo *repository.OrderRepository) error {
	config := sarama.NewConfig()
	config.ClientID = "go-kafka-delivery-consumer"
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.AutoCommit.Enable = false

	brokers := []string{"localhost:9092"}

	client, err := sarama.NewClient(brokers, config)
	if err != nil {
		slog.Error("Failed to create Sarama Client: %v", slog.Any("error", err))
		return err
	}
	defer client.Close()

	consumer, err := sarama.NewConsumer([]string{"localhost:9092"}, config)
	if err != nil {
		return err
	}
	defer consumer.Close()

	topic := "delivery-topic"

	offsetManager, err := sarama.NewOffsetManagerFromClient("order-group", client)
	if err != nil {
		return err
	}
	defer offsetManager.Close()

	partManager, err := offsetManager.ManagePartition(topic, 0)
	if err != nil {
		return err
	}
	defer partManager.Close()

	nextOffset, _ := partManager.NextOffset()
	if nextOffset < 0 {
		nextOffset = sarama.OffsetOldest
	}

	partConsumer, err := consumer.ConsumePartition("delivery-topic", 0, nextOffset)
	if err != nil {
		return err
	}
	defer partConsumer.Close()

	ctx := context.Background()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case msg := <-partConsumer.Messages():

			ProcessOrderMessage(msg, repo, ctx)

			partManager.MarkOffset(msg.Offset+1, "")
			offsetManager.Commit()

		case <-signals:
			slog.Info("Received os signal to stop")
			return nil

		}
	}
}

func ProcessOrderMessage(msg *sarama.ConsumerMessage, repo *repository.OrderRepository, ctx context.Context) error {
	slog.Debug("Received message:", "order json", string(msg.Value))

	var order model.Order
	if err := json.Unmarshal(msg.Value, &order); err != nil {
		slog.Error("Failed to unmarshal: ", slog.Any("error", err))
		return err
	}

	if err := ValidateOrderMessage(&order); err != nil {
		slog.Error("Failed to unmarshal: ", "error", slog.Any("error", err))
		return err
	}

	if err := repo.InsertOrder(ctx, &order); err != nil {
		slog.Error("Failed to insert: %v", slog.Any("error", err))
		return err
	}
	return nil
}

func ValidateOrderMessage(order *model.Order) error {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(order.Delivery.Email) {
		return errors.New("invalid email format")
	}
	phoneRegex := regexp.MustCompile(`^\+?[0-9]{10}$`)
	if !phoneRegex.MatchString(order.Delivery.Phone) {
		return errors.New("invalid phone format, should have 10 numbers in it")
	}
	return nil
}
