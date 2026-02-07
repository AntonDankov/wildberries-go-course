package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"wildberries-go-course/L3-3/database"
	"wildberries-go-course/L3-3/dto"
	"wildberries-go-course/L3-3/repository"

	"github.com/gin-gonic/gin"
)

func CreateComment(ctx context.Context, db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.CreateCommentDTO
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
			return
		}

		commentID, err := repository.AddComment(ctx, db, req.Text, req.ParentID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to add comment", "details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id": commentID,
		})
	}
}

func GetCommentWithReplies(ctx context.Context, db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		commentID, err := strconv.ParseInt(c.Param("commentID"), 10, 64)
		pageStr := c.Query("page")
		page := -1
		var e error
		if len(pageStr) != 0 {
			page, e = strconv.Atoi(pageStr)
			err = errors.Join(err, e)
		}
		pageSizeStr := c.Query("pageSize")
		pageSize := -1
		if len(pageSizeStr) != 0 {
			pageSize, e = strconv.Atoi(pageSizeStr)
			err = errors.Join(err, e)
		}
		maxDepthStr := c.Query("maxDepth")
		maxDepth := -1
		if len(maxDepthStr) != 0 {
			maxDepth, e = strconv.Atoi(maxDepthStr)
			err = errors.Join(err, e)
		}
		if err != nil {
			addJsonWithError(c, http.StatusBadRequest, err)
			return
		}
		comments, err := repository.GetCommentWithRepliesWithPaginationAndDepthLimit(ctx, db, commentID, page, pageSize, maxDepth)
		if err != nil {
			addJsonWithError(c, http.StatusInternalServerError, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"count":    len(comments),
			"comments": comments,
		})
	}
}

func DeleteCommentWithReplies(ctx context.Context, db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		commentID, err := strconv.ParseInt(c.Param("commentID"), 10, 64)
		if err != nil {
			addJsonWithError(c, http.StatusBadRequest, err)
			return
		}
		err = repository.DeleteCommentWithReplies(ctx, db, commentID)
		if err != nil {
			addJsonWithError(c, http.StatusInternalServerError, err)
			return
		}
	}
}

func FindCommentsByText(ctx context.Context, db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		searchText := c.Query("searchText")
		if len(searchText) == 0 {
			addJsonWithError(c, http.StatusBadRequest, fmt.Errorf("no seach text was provided, fill the searchText query parameter"))
			return
		}

		pageStr := c.Query("page")
		page := -1
		var err error
		var e error
		if len(pageStr) != 0 {
			page, e = strconv.Atoi(pageStr)
			err = errors.Join(err, e)
		}
		pageSizeStr := c.Query("pageSize")
		pageSize := -1
		if len(pageSizeStr) != 0 {
			pageSize, e = strconv.Atoi(pageSizeStr)
			err = errors.Join(err, e)
		}

		comments, err := repository.SearchInComments(ctx, db, searchText, page, pageSize)
		if err != nil {
			addJsonWithError(c, http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"count":    len(comments),
			"comments": comments,
		})
	}
}

func addJsonWithError(c *gin.Context, httpCode int, err error) {
	c.JSON(httpCode, gin.H{
		"error": err.Error(),
	})
}
