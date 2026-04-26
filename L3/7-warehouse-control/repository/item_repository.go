package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"wildberries-go-course/L3-7/database"
	"wildberries-go-course/L3-7/model"
)

func SetUserContext(ctx context.Context, db database.DBTX, userID int64) error {
	query := fmt.Sprintf("SET LOCAL app.current_user_id = %d", userID)

	_, err := db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to set user context: %v", err)
	}

	return nil
}

func CreateItem(ctx context.Context, db database.DBTX, ownerID int64, name string, price float64, amount int) (int64, error) {
	query := `
		INSERT INTO items (owner_id, name, price, amount) 
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var itemID int64
	err := db.QueryRowContext(ctx, query, ownerID, name, price, amount).Scan(&itemID)
	if err != nil {
		return 0, fmt.Errorf("failed to create item: %v", err)
	}

	return itemID, nil
}

func GetItem(ctx context.Context, db database.DBTX, itemID int64) (*model.Item, error) {
	query := `
		SELECT id, owner_id, name, price, amount, created_at, updated_at 
		FROM items 
		WHERE id = $1
	`

	var item model.Item
	err := db.QueryRowContext(ctx, query, itemID).Scan(
		&item.ID,
		&item.OwnerID,
		&item.Name,
		&item.Price,
		&item.Amount,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("item not found: %d", itemID)
		}
		return nil, fmt.Errorf("failed to get item: %v", err)
	}

	return &item, nil
}

func GetAllItems(ctx context.Context, db database.DBTX) ([]model.Item, error) {
	query := `
		SELECT id, owner_id, name, price, amount, created_at, updated_at 
		FROM items 
		ORDER BY created_at DESC
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get items: %v", err)
	}
	defer rows.Close()

	var items []model.Item
	for rows.Next() {
		var item model.Item
		err := rows.Scan(
			&item.ID,
			&item.OwnerID,
			&item.Name,
			&item.Price,
			&item.Amount,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan item: %v", err)
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating items: %v", err)
	}

	return items, nil
}

func UpdateItem(ctx context.Context, db database.DBTX, itemID int64, name string, price float64, amount int, ownerID int64, role model.RoleType) error {
	baseQuery := `
		UPDATE items 
		SET name = $1, price = $2, amount = $3, updated_at = NOW() 
	`

	conditions := []string{}
	args := []any{name, price, amount}
	sqlArgIndex := 4
	{
		conditions = append(conditions, fmt.Sprintf("id = $%d", sqlArgIndex))
		args = append(args, itemID)
		sqlArgIndex++
	}

	if (role & (model.Admin | model.Manager)) == 0 {
		conditions = append(conditions, fmt.Sprintf("owner_id = $%d", sqlArgIndex))
		args = append(args, ownerID)
		sqlArgIndex++
	}

	query := baseQuery
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update item: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("item not found: %d", itemID)
	}

	return nil
}

func DeleteItem(ctx context.Context, db database.DBTX, itemID int64, userID int64) error {
	query := `DELETE FROM items WHERE id = $1 and owner_id = $2`

	result, err := db.ExecContext(ctx, query, itemID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete item: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("item not found: %d", itemID)
	}

	return nil
}

func GetItemHistory(ctx context.Context, db database.DBTX, itemID int64) ([]model.ItemHistory, error) {
	query := `
		SELECT h.id, h.item_id, h.name, h.price, h.amount, h.action, h.changed_by, u.name, h.changed_at 
		FROM item_history h
		join users u on h.changed_by=u.id
		WHERE item_id = $1 
		ORDER BY changed_at DESC
	`

	rows, err := db.QueryContext(ctx, query, itemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get item history: %v", err)
	}
	defer rows.Close()

	var history []model.ItemHistory
	for rows.Next() {
		var h model.ItemHistory
		err := rows.Scan(
			&h.ID,
			&h.ItemID,
			&h.Name,
			&h.Price,
			&h.Amount,
			&h.Action,
			&h.UserID,
			&h.Username,
			&h.ChangedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan history: %v", err)
		}
		history = append(history, h)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating history: %v", err)
	}

	return history, nil
}
