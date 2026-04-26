package repository

import (
	"context"
	"database/sql"
	"fmt"
	"wildberries-go-course/L3-4/database"
	"wildberries-go-course/L3-4/model"

	"github.com/disintegration/imaging"
)

func AddImageProcess(ctx context.Context, db *database.Database, imageID string, extension imaging.Format) error {
	query := `
		INSERT INTO image_process (id, extension, process_type) 
		VALUES ($1, $2, $3)
	`
	_, err := db.Master.ExecContext(ctx, query, imageID, extension.String(), model.Waiting)
	if err != nil {
		return fmt.Errorf("failed to add image process: %v", err)
	}

	return nil
}

func DeleteImageProcess(ctx context.Context, db *database.Database, imageID string) error {
	query := `
		UPDATE image_process 
		SET process_type = $1, updated = NOW() 
		WHERE id = $2
	`
	result, err := db.Master.ExecContext(ctx, query, model.Deleted, imageID)
	if err != nil {
		return fmt.Errorf("failed to delete image process: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("image process not found: %s", imageID)
	}

	return nil
}

func GetImageProcess(ctx context.Context, db *database.Database, imageID string) (*model.ImageStatus, error) {
	query := `
		SELECT id, extension, process_type 
		FROM image_process 
		WHERE id = $1
	`

	var imageProcess model.ImageStatus
	err := db.Master.QueryRowContext(ctx, query, imageID).Scan(
		&imageProcess.ID,
		&imageProcess.Extension,
		&imageProcess.ProcessType,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("image process not found: %s", imageID)
		}
		return nil, fmt.Errorf("failed to get image process: %v", err)
	}

	return &imageProcess, nil
}

func UpdateImageProcess(ctx context.Context, db *database.Database, imageID string, processType model.ImageProcessingType) error {
	query := `
		UPDATE image_process 
		SET process_type = $1, updated = NOW() 
		WHERE id = $2
	`
	_, err := db.Master.ExecContext(ctx, query, processType, imageID)
	if err != nil {
		return fmt.Errorf("failed to update image process type: %v", err)
	}

	return nil
}

func GetImagesStatusWithPagination(ctx context.Context, db *database.Database, page int, pageSize int) ([]model.ImageStatus, int64, error) {
	query := `
		SELECT id, extension, process_type, COUNT(*) OVER() as total_count
		FROM image_process 
		WHERE process_type != $1
	ORDER BY created DESC
		LIMIT CASE WHEN $2 = -1 THEN NULL ELSE $2 END
			OFFSET CASE WHEN $3 = -1 THEN 0 ELSE $3 END
	`
	offset := 0
	if page != -1 && pageSize != -1 {
		offset = page * pageSize
	}
	rows, err := db.Master.QueryContext(ctx, query, model.Deleted, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var listImageProcess []model.ImageStatus
	var totalCount int64
	for rows.Next() {
		var imageProcess model.ImageStatus
		if err := rows.Scan(&imageProcess.ID, &imageProcess.Extension, &imageProcess.ProcessType, &totalCount); err != nil {
			return nil, 0, err
		}
		listImageProcess = append(listImageProcess, imageProcess)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return listImageProcess, totalCount, nil
}
