package broker

import (
	"bytes"
	"context"
	"encoding/gob"
	"log"
	"os"
	"time"
	model "widlberries-go-course/L3-1/model"
	"widlberries-go-course/L3-1/repository"

	"github.com/rabbitmq/amqp091-go"
	"github.com/wb-go/wbf/rabbitmq"
	"github.com/wb-go/wbf/retry"
)

var GlobalRabbitClient *rabbitmq.RabbitClient

const (
	MaxAttempts  = 2
	MaxDelay     = time.Hour * 8
	ExchangeName = "notifications_exchange"
	QueueName    = "notifications_queue"
	RoutingKey   = "notifications_send"
)

func InitRabbitMQ() error {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	defaultTime := time.Second * 7
	defaultStrategy := retry.Strategy{
		Attempts: 3,
		Delay:    time.Second * 2,
		Backoff:  2.0,
	}
	rabbitConfig := rabbitmq.ClientConfig{
		URL:            rabbitURL,
		ConnectTimeout: defaultTime,
		Heartbeat:      defaultTime,
		PublishRetry:   defaultStrategy,
		ConsumeRetry:   defaultStrategy,
	}
	rabbitClient, err := rabbitmq.NewClient(rabbitConfig)
	if err != nil {
		return err
	}
	err = rabbitClient.DeclareExchange(
		ExchangeName,
		"x-delayed-message",
		true,
		false,
		false,
		map[string]any{
			"x-delayed-type": "direct",
		},
	)
	if err != nil {
		return err
	}

	ch, err := rabbitClient.GetChannel()
	if err != nil {
		return err
	}

	_, err = ch.QueueDeclare(
		QueueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	err = ch.QueueBind(
		QueueName,
		RoutingKey,
		ExchangeName,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	GlobalRabbitClient = rabbitClient

	return nil
}

func PublishNotificaitonMessage(rabbitClient *rabbitmq.RabbitClient, message NotificationMessage, delay time.Duration) error {
	ch, err := rabbitClient.GetChannel()
	if err != nil {
		return err
	}

	defer ch.Close()

	headers := make(map[string]any)
	rabbitDelay := max(int64(delay/time.Millisecond), 0)
	headers["x-delay"] = rabbitDelay
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(message); err != nil {
		return err
	}
	messageBytes := buf.Bytes()

	err = ch.PublishWithContext(
		ctx,
		ExchangeName,
		RoutingKey,
		false,
		false,
		amqp091.Publishing{
			ContentType:  "application/octet-stream",
			Body:         messageBytes,
			DeliveryMode: 2,
			Headers:      headers,
		},
	)
	if err != nil {
		return err
	}

	log.Printf("Published message: %v with delay %d \n", message, rabbitDelay)
	return nil
}

func StartConsumer(rabbitClient *rabbitmq.RabbitClient, rep repository.NotificationRepository) error {
	ch, err := rabbitClient.GetChannel()
	if err != nil {
		return err
	}

	err = ch.Qos(1, 0, false)
	if err != nil {
		return err
	}

	queue, err := ch.Consume(
		QueueName,
		"notification-worker",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	log.Println("Consumer started")

	for msg := range queue {
		var message NotificationMessage
		buf := bytes.NewBuffer(msg.Body)
		decoder := gob.NewDecoder(buf)
		err := decoder.Decode(&message)
		if err != nil {
			log.Printf("Failed to decode message: %v\n", err)
			msg.Nack(false, false)
			continue
		}

		log.Printf("Received notification: ID=%d\n", message.ID)

		err = processMessage(rabbitClient, message, rep)
		if err != nil {
			// it might have been better to use transaction for db
			// and dont commit if the error is received
			msg.Nack(false, true)
		} else {
			msg.Ack(false)
		}

	}

	return nil
}

func sendNotification(notification model.Notification) error {
	log.Printf("Notification: %d %s \n", notification.ID, notification.Text)
	return nil
}

func processMessage(rabbitClient *rabbitmq.RabbitClient, message NotificationMessage, rep repository.NotificationRepository) error {
	ctx := context.Background()
	notification, err := rep.GetNotification(ctx, message.ID)
	if err != nil {
		return err
	}
	if notification.Status == model.Deleted {
		log.Println("Notification doesn't exists anymore")
		return nil
	}
	err = sendNotification(*notification)
	if err == nil {
		return rep.UpdateNotificationStatus(ctx, message.ID, model.Sended)
	}
	// If we failed to send notification
	// reschedule message with delay if its less then MaxAttempts
	isResend := recalculateNotification(notification)
	if err := rep.UpdateNotification(ctx, *notification); err != nil {
		return err
	}
	if isResend {
		if err := PublishNotificaitonMessage(rabbitClient, message, time.Until(notification.NextAttemptAt)); err != nil {
			return err
		}
	}
	return nil
}

func recalculateNotification(notification *model.Notification) bool {
	if notification.FailedAttempts >= MaxAttempts {
		log.Printf("Notification %d failed\n", notification.ID)
		notification.Status = model.Failed
		return false
	}
	notification.FailedAttempts++
	delay := calculateDelay(notification.FailedAttempts)
	notification.Status = model.FailedAndAwaits
	notification.NextAttemptAt = time.Now().Add(delay)
	return true
}

// Attempts should be at least 1
func calculateDelay(attempt int) time.Duration {
	delay := time.Minute * time.Duration(1<<uint(attempt-1))
	return min(delay, MaxDelay)
}
