package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"wildberries-go-course/L3-4/database"
	"wildberries-go-course/L3-4/kafka"
	"wildberries-go-course/L3-4/model"
	"wildberries-go-course/L3-4/repository"
	"wildberries-go-course/L3-4/storage"
	"wildberries-go-course/L3-4/util"

	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/wb-go/wbf/zlog"
)

func UploadImage(ctx context.Context, db *database.Database, imageStorage *storage.ImageStorage, imageProducer *kafka.ImageBrokerProducer) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, header, err := c.Request.FormFile("image")
		if err != nil {
			addJSONWithError(c, http.StatusBadRequest, fmt.Errorf("invalid file: %w", err))
			return
		}
		defer file.Close()

		ext, err := imaging.FormatFromFilename(header.Filename)
		if err != nil {
			addJSONWithError(c, http.StatusBadRequest, fmt.Errorf("unsupported format: %s", ext))
			return
		}
		imageData, err := io.ReadAll(file)
		if err != nil {
			addJSONWithError(c, http.StatusInternalServerError, fmt.Errorf("failed to read file: %w", err))
			return
		}

		var operations model.ImageOperations
		var resizeWidth, resizeHeight int

		if c.PostForm("miniature") == "true" {
			operations |= model.Resize
			resizeWidth = 150
			resizeHeight = 150
		} else if c.PostForm("resize") == "true" {
			operations |= model.Resize
			resizeWidth, _ = strconv.Atoi(c.PostForm("resize_width"))
			resizeHeight, _ = strconv.Atoi(c.PostForm("resize_height"))
			if resizeWidth <= 0 || resizeHeight <= 0 {
				addJSONWithError(c, http.StatusBadRequest, fmt.Errorf("invalid resize dimensions"))
				return
			}
		}

		if c.PostForm("watermark") == "true" {
			operations |= model.Watermark
		}

		imageID := strings.ReplaceAll(uuid.NewString(), "-", "")

		image := model.ImageInfo{
			ID:           imageID,
			Type:         model.Original,
			Extension:    ext,
			Operations:   operations,
			ResizeWidth:  resizeWidth,
			ResizeHeight: resizeHeight,
		}

		if err := repository.AddImageProcess(ctx, db, imageID, ext); err != nil {
			addJSONWithError(c, http.StatusInternalServerError, fmt.Errorf("failed to store image processing to database: %w", err))
			return
		}
		// change it to store in kafka
		if err := imageStorage.StoreImage(image, imageData); err != nil {
			addJSONWithError(c, http.StatusInternalServerError, fmt.Errorf("failed to store image: %w", err))
			return
		}
		zlog.Logger.Info().Msg("Going to send image to kafka")
		if err := imageProducer.SendImage(ctx, &image); err != nil {
			addJSONWithError(c, http.StatusInternalServerError, fmt.Errorf("failed to store image processing to database: %w", err))
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"id":     imageID,
			"status": "processing",
		})
	}
}

func GetImage(ctx context.Context, db *database.Database, imageStorage *storage.ImageStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		imageID := c.Param("id")
		if imageID == "" {
			addJSONWithError(c, http.StatusBadRequest, fmt.Errorf("missing image ID"))
			return
		}

		imageProcess, err := repository.GetImageProcess(ctx, db, imageID)
		if err != nil {
			addJSONWithError(c, http.StatusInternalServerError, fmt.Errorf("failed to get info about image processing: %w", err))
			return
		}

		if imageProcess.ProcessType == model.Waiting {
			addJSONWithError(c, http.StatusAccepted, fmt.Errorf("image still in queue and waiting to be processed"))
			return
		}

		if imageProcess.ProcessType == model.Deleted {
			addJSONWithError(c, http.StatusNotFound, fmt.Errorf("image was deleted"))
			return
		}

		if imageProcess.ProcessType == model.Failed {
			addJSONWithError(c, http.StatusInternalServerError, fmt.Errorf("image processing failed, check your image and reupload again"))
			return
		}

		ext, _ := imaging.FormatFromExtension(imageProcess.Extension)

		image := model.ImageInfo{
			ID:        imageID,
			Type:      model.Modified,
			Extension: ext,
		}

		imageData, err := imageStorage.GetImageData(&image)
		if err != nil {
			addJSONWithError(c, http.StatusNotFound, fmt.Errorf("image not found: %w", err))
			return
		}

		contentType := util.FormatToContentType(image.Extension)

		c.Data(http.StatusOK, contentType, imageData)
	}
}

func getImagesStatus(ctx context.Context, db *database.Database, imageStorage *storage.ImageStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
	}
}

// DeleteImage only marking as deleted
func DeleteImage(ctx context.Context, db *database.Database, imageStorage *storage.ImageStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		imageID := c.Param("id")
		if imageID == "" {
			addJSONWithError(c, http.StatusBadRequest, fmt.Errorf("missing image ID"))
			return
		}

		err := repository.DeleteImageProcess(ctx, db, imageID)
		if err != nil {
			addJSONWithError(c, http.StatusInternalServerError, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "image deleted successfully",
		})
	}
}

func GetImagesStatus(ctx context.Context, db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
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
		if err != nil {
			addJSONWithError(c, http.StatusBadRequest, err)
			return
		}
		listImageProcess, err := repository.GetImagesStatusWithPagination(ctx, db, page, pageSize)
		if err != nil {
			addJSONWithError(c, http.StatusInternalServerError, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"count":    len(listImageProcess),
			"comments": listImageProcess,
		})
	}
}

func addJSONWithError(c *gin.Context, httpCode int, err error) {
	c.JSON(httpCode, gin.H{
		"error": err.Error(),
	})
}
