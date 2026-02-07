package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"wildberries-go-course/L3-7/auth"
	"wildberries-go-course/L3-7/database"
	"wildberries-go-course/L3-7/dto"
	"wildberries-go-course/L3-7/model"
	"wildberries-go-course/L3-7/repository"

	"github.com/wb-go/wbf/ginext"
)

func CreateItem(ctx context.Context, db *database.Database) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		userClaims, err := auth.Authorize(ctx, db, c, model.Admin|model.Manager|model.Owner)
		if err != nil {
			addJSONWithError(c, http.StatusUnauthorized, err)
			return
		}
		var itemDTO dto.ItemDTO

		if err := c.ShouldBindJSON(&itemDTO); err != nil {
			addJSONWithError(c, http.StatusBadRequest, fmt.Errorf("invalid request: %w", err))
			return
		}

		tx, err := db.Master.BeginTx(ctx, nil)
		if err != nil {
			addJSONWithError(c, http.StatusInternalServerError, fmt.Errorf("failed to begin transaction: %w", err))
			return
		}
		defer tx.Rollback()

		if err := repository.SetUserContext(ctx, tx, userClaims.UserID); err != nil {
			addJSONWithError(c, http.StatusInternalServerError, fmt.Errorf("failed to set user context: %w", err))
			return
		}

		itemID, err := repository.CreateItem(ctx, tx, userClaims.UserID, itemDTO.Name, itemDTO.Price, itemDTO.Amount)
		if err != nil {
			addJSONWithError(c, http.StatusInternalServerError, fmt.Errorf("failed to create item: %w", err))
			return
		}

		if err := tx.Commit(); err != nil {
			addJSONWithError(c, http.StatusInternalServerError, fmt.Errorf("failed to commit transaction: %w", err))
			return
		}

		itemDTO.ID = itemID
		c.JSON(http.StatusCreated, itemDTO)
	}
}

func GetItems(ctx context.Context, db *database.Database) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		items, err := repository.GetAllItems(ctx, db.Master)
		if err != nil {
			addJSONWithError(c, http.StatusInternalServerError, fmt.Errorf("failed to get items: %w", err))
			return
		}

		c.JSON(http.StatusOK, ginext.H{
			"items": items,
			"size":  len(items),
		})
	}
}

func GetItem(ctx context.Context, db *database.Database) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		itemIDStr := c.Param("id")
		if itemIDStr == "" {
			addJSONWithError(c, http.StatusBadRequest, fmt.Errorf("missing item ID"))
			return
		}

		itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
		if err != nil {
			addJSONWithError(c, http.StatusBadRequest, fmt.Errorf("invalid item ID: %w", err))
			return
		}

		item, err := repository.GetItem(ctx, db.Master, itemID)
		if err != nil {
			addJSONWithError(c, http.StatusNotFound, fmt.Errorf("failed to get item: %w", err))
			return
		}

		c.JSON(http.StatusOK, ginext.H{
			"item": item,
		})
	}
}

func UpdateItem(ctx context.Context, db *database.Database) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		userClaims, err := auth.Authorize(ctx, db, c, model.Admin|model.Manager|model.Owner)
		if err != nil {
			addJSONWithError(c, http.StatusUnauthorized, err)
			return
		}
		itemIDStr := c.Param("id")
		if itemIDStr == "" {
			addJSONWithError(c, http.StatusBadRequest, fmt.Errorf("missing item ID"))
			return
		}

		itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
		if err != nil {
			addJSONWithError(c, http.StatusBadRequest, fmt.Errorf("invalid item ID: %w", err))
			return
		}

		var itemFromRequest model.Item

		if err := c.ShouldBindJSON(&itemFromRequest); err != nil {
			addJSONWithError(c, http.StatusBadRequest, fmt.Errorf("invalid request: %w", err))
			return
		}

		tx, err := db.Master.BeginTx(ctx, nil)
		if err != nil {
			addJSONWithError(c, http.StatusInternalServerError, fmt.Errorf("failed to begin transaction: %w", err))
			return
		}
		defer tx.Rollback()

		if err := repository.SetUserContext(ctx, tx, userClaims.UserID); err != nil {
			addJSONWithError(c, http.StatusInternalServerError, fmt.Errorf("failed to set user context: %w", err))
			return
		}

		if err := repository.UpdateItem(ctx, tx, itemID, itemFromRequest.Name, itemFromRequest.Price, itemFromRequest.Amount, userClaims.UserID); err != nil {
			addJSONWithError(c, http.StatusInternalServerError, fmt.Errorf("failed to update item: %w", err))
			return
		}

		if err := tx.Commit(); err != nil {
			addJSONWithError(c, http.StatusInternalServerError, fmt.Errorf("failed to commit transaction: %w", err))
			return
		}

		c.JSON(http.StatusOK, ginext.H{
			"message": "item updated successfully",
		})
	}
}

func DeleteItem(ctx context.Context, db *database.Database) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		userClaims, err := auth.Authorize(ctx, db, c, model.Admin|model.Manager|model.Owner)
		if err != nil {
			addJSONWithError(c, http.StatusUnauthorized, err)
			return
		}
		itemIDStr := c.Param("id")
		if itemIDStr == "" {
			addJSONWithError(c, http.StatusBadRequest, fmt.Errorf("missing item ID"))
			return
		}

		itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
		if err != nil {
			addJSONWithError(c, http.StatusBadRequest, fmt.Errorf("invalid item ID: %w", err))
			return
		}

		tx, err := db.Master.BeginTx(ctx, nil)
		if err != nil {
			addJSONWithError(c, http.StatusInternalServerError, fmt.Errorf("failed to begin transaction: %w", err))
			return
		}
		defer tx.Rollback()

		if err := repository.SetUserContext(ctx, tx, userClaims.UserID); err != nil {
			addJSONWithError(c, http.StatusInternalServerError, fmt.Errorf("failed to set user context: %w", err))
			return
		}

		if err := repository.DeleteItem(ctx, tx, itemID, userClaims.UserID); err != nil {
			addJSONWithError(c, http.StatusInternalServerError, fmt.Errorf("failed to delete item: %w", err))
			return
		}

		if err := tx.Commit(); err != nil {
			addJSONWithError(c, http.StatusInternalServerError, fmt.Errorf("failed to commit transaction: %w", err))
			return
		}

		c.JSON(http.StatusOK, ginext.H{
			"message": "item deleted successfully",
		})
	}
}

func GetItemHistory(ctx context.Context, db *database.Database) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		_, err := auth.Authorize(ctx, db, c, model.Admin|model.Manager)
		if err != nil {
			addJSONWithError(c, http.StatusUnauthorized, err)
			return
		}
		itemIDStr := c.Param("id")
		if itemIDStr == "" {
			addJSONWithError(c, http.StatusBadRequest, fmt.Errorf("missing item ID"))
			return
		}

		itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
		if err != nil {
			addJSONWithError(c, http.StatusBadRequest, fmt.Errorf("invalid item ID: %w", err))
			return
		}

		history, err := repository.GetItemHistory(ctx, db.Master, itemID)
		if err != nil {
			addJSONWithError(c, http.StatusInternalServerError, fmt.Errorf("failed to get history: %w", err))
			return
		}

		c.JSON(http.StatusOK, ginext.H{
			"history": history,
			"size":    len(history),
		})
	}
}
