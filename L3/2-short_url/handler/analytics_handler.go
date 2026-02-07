package handler

import (
	"context"
	"net/http"
	"time"
	database "wildberries-go-course/L3-2/database"
	repository "wildberries-go-course/L3-2/repository"
	util "wildberries-go-course/L3-2/util"

	"github.com/gin-gonic/gin"
)

func GetAnalyticsFull(ctx context.Context, db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		shortURL := c.Param("shortURL")
		id, err := util.DecodeBase58(shortURL)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid short url", "details": err.Error()})
			return
		}

		analytics, err := repository.GetAnalyticsFull(ctx, db, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get analytics", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"count":     len(analytics),
			"analytics": analytics,
		})
	}
}

func GetAnalyticsByDay(ctx context.Context, db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		shortURL := c.Param("shortURL")
		id, err := util.DecodeBase58(shortURL)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid short url", "details": err.Error()})
			return
		}

		startDateStr := c.Query("start_date")
		endDateStr := c.Query("end_date")

		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format", "details": "Expected format: YYYY-MM-DD"})
			return
		}

		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format", "details": "Expected format: YYYY-MM-DD"})
			return
		}

		analytics, err := repository.GetAnalyticsAggreatedByDay(ctx, db, id, startDate, endDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get analytics", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"count":     len(analytics),
			"analytics": analytics,
		})
	}
}

func GetAnalyticsByMonth(ctx context.Context, db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		shortURL := c.Param("shortURL")
		id, err := util.DecodeBase58(shortURL)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid short url", "details": err.Error()})
			return
		}

		startDateStr := c.Query("start_date")
		endDateStr := c.Query("end_date")

		startDate, err := time.Parse("2006-01", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format", "details": "Expected format: YYYY-MM"})
			return
		}

		endDate, err := time.Parse("2006-01", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format", "details": "Expected format: YYYY-MM"})
			return
		}

		analytics, err := repository.GetAnalyticsAggregatedByMonth(ctx, db, id, startDate, endDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get analytics", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"count":     len(analytics),
			"analytics": analytics,
		})
	}
}

func GetAnalyticsByUserAgent(ctx context.Context, db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		shortURL := c.Param("shortURL")
		id, err := util.DecodeBase58(shortURL)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid short url", "details": err.Error()})
			return
		}

		startDateStr := c.Query("start_date")
		endDateStr := c.Query("end_date")

		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format", "details": "Expected format: YYYY-MM-DD"})
			return
		}

		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format", "details": "Expected format: YYYY-MM-DD"})
			return
		}

		analytics, err := repository.GetAnalyticsAggregatedByUserAgent(ctx, db, id, startDate, endDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get analytics", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"count":     len(analytics),
			"analytics": analytics,
		})
	}
}
