package repository

import (
	"context"
	"fmt"

	database "wildberries-go-course/L3-2/database"
	model "wildberries-go-course/L3-2/model"
)

func AddShortURL(ctx context.Context, db *database.Database, shortUrl model.ShortUrl) (int64, error) {
	query := `INSERT INTO short_url (url,created_at)
		VALUES ($1,$2)
	RETURNING id
	`
	var id int64
	err := db.Master.QueryRowContext(ctx, query,
		shortUrl.Url,
		shortUrl.CreatedAt,
	).Scan(&id)
	if err != nil {
		return -1, fmt.Errorf("failed to add ShortUrl: %v", err)
	}
	return id, nil
}

func GetURLByID(ctx context.Context, db *database.Database, id int64) (string, error) {
	query := `SELECT url FROM short_url
	WHERE id = $1
	`
	var url string
	err := db.Master.QueryRowContext(ctx, query,
		id,
	).Scan(&url)
	if err != nil {
		return "", fmt.Errorf("failed to get ShortUrl: %v", err)
	}
	return url, nil
}
