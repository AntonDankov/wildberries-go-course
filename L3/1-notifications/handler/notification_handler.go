package handler

import (
	"net/http"
	"strconv"
	"time"
	dto "widlberries-go-course/L3-1/dto"
	model "widlberries-go-course/L3-1/model"
	"widlberries-go-course/L3-1/repository"

	broker "widlberries-go-course/L3-1/broker"

	"github.com/gin-gonic/gin"
	"github.com/wb-go/wbf/rabbitmq"
)

func CreateNotificationHandler(r repository.NotificationRepository, rabbitClient *rabbitmq.RabbitClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.CreateNotificationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
			return
		}

		notification := model.Notification{
			Text:           req.Text,
			Status:         model.AwaitsSending,
			CreatedAt:      time.Now(),
			SendAt:         req.SendAt,
			FailedAttempts: 0,
		}

		id, err := r.AddNotification(c.Request.Context(), notification)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		message := broker.NotificationMessage{
			ID: id,
		}
		delay := time.Until(req.SendAt)

		err = broker.PublishNotificaitonMessage(rabbitClient, message, delay)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"id": id,
		})
	}
}

func GetNotificationHandler(r repository.NotificationRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "id not valid"})
			return
		}

		notification, err := r.GetNotification(c.Request.Context(), id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if notification.Status == model.Deleted {
			c.JSON(http.StatusNotFound, gin.H{"error": "notification doesn't exist"})
			return
		}
		c.JSON(http.StatusOK, notification)
	}
}

func DeleteNotificationHandler(r repository.NotificationRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "id not valid"})
			return
		}

		err = r.UpdateNotificationStatus(c.Request.Context(), id, model.Deleted)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "deleted notification"})
	}
}
