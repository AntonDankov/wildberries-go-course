package handler

import (
	"context"
	"net/http"
	"time"

	database "wildberries-go-course/L3-2/database"
	dto "wildberries-go-course/L3-2/dto"
	model "wildberries-go-course/L3-2/model"
	repository "wildberries-go-course/L3-2/repository"
	util "wildberries-go-course/L3-2/util"

	"github.com/gin-gonic/gin"
)

func CreateShortLink(ctx context.Context, db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var urlDTO dto.ShortUrlDTO
		if err := c.ShouldBindJSON(&urlDTO); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
			return
		}
		shortUrl := model.ShortUrl{
			Url:       urlDTO.Url,
			CreatedAt: time.Now(),
		}
		id, err := repository.AddShortURL(ctx, db, shortUrl)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create short link", "details": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{
			"short_url": util.EncodeBase58(id),
		})
	}
}

func VisitShortLink(ctx context.Context, db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		shortURL := c.Param("shortURL")
		id, err := util.DecodeBase58(shortURL)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid short url", "details": err.Error()})
			return
		}
		url, err := repository.GetURLByID(ctx, db, id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid short url", "details": err.Error()})
			return
		}
		userAgent := c.GetHeader("User-Agent")
		analytic := model.Analytic{
			UserAgent: userAgent,
			VisitTime: time.Now(),
			URLID:     id,
		}
		_, err = repository.AddAnalytic(ctx, db, analytic)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to add analytics", "details": err.Error()})
			return
		}
		c.Redirect(http.StatusFound, url)
	}
}
