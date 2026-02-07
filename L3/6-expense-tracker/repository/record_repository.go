package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"wildberries-go-course/L3-6/database"
	"wildberries-go-course/L3-6/dto"
	"wildberries-go-course/L3-6/model"
)

func CreateRecord(ctx context.Context, db database.DBTX, record *model.Record) (int64, error) {
	query := `
		INSERT INTO record (type, category, amount, date) 
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var recordID int64
	err := db.QueryRowContext(ctx, query, record.Type, record.Category, record.Amount, record.Date).Scan(&recordID)
	if err != nil {
		return 0, fmt.Errorf("failed to create record: %v", err)
	}

	return recordID, nil
}

func GetRecord(ctx context.Context, db database.DBTX, recordID int64) (*model.Record, error) {
	query := `
		SELECT id, type, category, amount, date 
		FROM record 
		WHERE id = $1
	`

	var record model.Record
	err := db.QueryRowContext(ctx, query, recordID).Scan(
		&record.ID,
		&record.Type,
		&record.Category,
		&record.Amount,
		&record.Date,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("record not found: %d", recordID)
		}
		return nil, fmt.Errorf("failed to get record: %v", err)
	}

	return &record, nil
}

func UpdateRecord(ctx context.Context, db database.DBTX, recordID int64, record *model.Record) error {
	query := `
		UPDATE record 
		SET type = $1, category = $2, amount = $3, date = $4 
		WHERE id = $5
	`

	result, err := db.ExecContext(ctx, query, record.Type, record.Category, record.Amount, record.Date, recordID)
	if err != nil {
		return fmt.Errorf("failed to update record: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("record not found: %d", recordID)
	}

	return nil
}

func DeleteRecord(ctx context.Context, db database.DBTX, recordID int64) error {
	query := `
		DELETE FROM record 
		WHERE id = $1
	`

	result, err := db.ExecContext(ctx, query, recordID)
	if err != nil {
		return fmt.Errorf("failed to delete record: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("record not found: %d", recordID)
	}

	return nil
}

func GetRecords(ctx context.Context, db database.DBTX, filter *dto.AnalyticsFilter) ([]*model.Record, error) {
	query := "SELECT id, type, category, amount, date FROM record"

	conditions := []string{}
	args := []interface{}{}
	sqlArgIndex := 1

	if !filter.From.IsZero() {
		conditions = append(conditions, fmt.Sprintf("date >= $%d", sqlArgIndex))
		args = append(args, filter.From)
		sqlArgIndex++
	}

	if !filter.To.IsZero() {
		conditions = append(conditions, fmt.Sprintf("date <= $%d", sqlArgIndex))
		args = append(args, filter.To)
		sqlArgIndex++
	}

	if filter.Type != nil {
		conditions = append(conditions, fmt.Sprintf("type = $%d", sqlArgIndex))
		args = append(args, *filter.Type)
		sqlArgIndex++
	}

	if filter.Category != nil {
		conditions = append(conditions, fmt.Sprintf("category = $%d", sqlArgIndex))
		args = append(args, *filter.Category)
		sqlArgIndex++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY date DESC"

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get records: %v", err)
	}
	defer rows.Close()

	var records []*model.Record
	for rows.Next() {
		var record model.Record
		err := rows.Scan(
			&record.ID,
			&record.Type,
			&record.Category,
			&record.Amount,
			&record.Date,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan record: %v", err)
		}
		records = append(records, &record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating records: %v", err)
	}

	return records, nil
}
