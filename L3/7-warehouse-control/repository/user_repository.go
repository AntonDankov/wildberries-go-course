package repository

import (
	"context"
	"database/sql"
	"fmt"
	"wildberries-go-course/L3-7/database"
	"wildberries-go-course/L3-7/model"
)

func CreateUser(ctx context.Context, db database.DBTX, name string, passwordHash string, role model.RoleType) (int64, error) {
	query := `
		INSERT INTO users (name, password_hash, role) 
		VALUES ($1, $2, $3)
		RETURNING id
	`

	var userID int64
	err := db.QueryRowContext(ctx, query, name, passwordHash, role).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("failed to create user: %v", err)
	}

	return userID, nil
}

func GetUserByName(ctx context.Context, db database.DBTX, name string) (*model.User, error) {
	query := `
		SELECT id, name, password_hash, role 
		FROM users 
		WHERE name = $1
	`

	var user model.User
	err := db.QueryRowContext(ctx, query, name).Scan(
		&user.ID,
		&user.Name,
		&user.Password,
		&user.Role,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %s", name)
		}
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	return &user, nil
}

func GetUserByID(ctx context.Context, db database.DBTX, userID int64) (*model.User, error) {
	query := `
		SELECT id, name, password_hash, role 
		FROM users 
		WHERE id = $1
	`

	var user model.User
	err := db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.Name,
		&user.Password,
		&user.Role,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %d", userID)
		}
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	return &user, nil
}
