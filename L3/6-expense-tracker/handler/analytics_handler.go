package handler

import (
	"context"
	"fmt"
	"net/http"
	"wildberries-go-course/L3-6/database"
	"wildberries-go-course/L3-6/dto"
	"wildberries-go-course/L3-6/repository"

	"github.com/wb-go/wbf/ginext"
)

func GetAnalytics(ctx context.Context, db *database.Database) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		filter := dto.ParseFilter(c)

		analytics, err := repository.GetAnalytics(ctx, db.Master, filter)
		if err != nil {
			addJSONWithError(c, http.StatusInternalServerError, fmt.Errorf("failed to get analytics: %w", err))
			return
		}

		c.JSON(http.StatusOK, analytics)
	}
}
