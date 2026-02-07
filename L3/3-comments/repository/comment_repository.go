package repository

import (
	"context"
	"fmt"
	"wildberries-go-course/L3-3/database"
	"wildberries-go-course/L3-3/dto"
)

func AddComment(ctx context.Context, db *database.Database, text string, parentID int64) (int64, error) {
	tx, err := db.Master.BeginTx(ctx, nil)
	if err != nil {
		return -1, fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	var commentID int64

	query := `
		INSERT INTO comment (text) VALUES ($1) RETURNING id
	`
	err = tx.QueryRowContext(ctx, query, text).Scan(&commentID)
	if err != nil {
		return -1, fmt.Errorf("failed to add comment: %v", err)
	}

	var closureQuery string
	if parentID == -1 {
		closureQuery = `
			INSERT INTO comment_closure_tree (parent_id, child_id, depth) 
		VALUES ($1,$1,0)
		`
		_, err = tx.ExecContext(ctx, closureQuery, commentID)
	} else {
		closureQuery = `
			INSERT INTO comment_closure_tree (parent_id, child_id, depth)
		SELECT parent_id, $1::INTEGER, depth + 1 FROM comment_closure_tree
			WHERE child_id = $2 UNION ALL
		SELECT $1::INTEGER,$1::INTEGER,0
		`
		_, err = tx.ExecContext(ctx, closureQuery, commentID, parentID)
	}
	if err != nil {
		return -1, fmt.Errorf("failed to add comment path: %v", err)
	}

	if err = tx.Commit(); err != nil {
		return -1, fmt.Errorf("failed to commit transaction with adding comment: %v", err)
	}

	return commentID, nil
}

func GetCommentWithRepliesWithPaginationAndDepthLimit(ctx context.Context, db *database.Database, commentID int64, page int, pageSize int, maxDepth int) ([]dto.CommentDTO, error) {
	query := `
	WITH root_comment AS (
			SELECT 
				c.id, c.text, ct.depth,
				-1 as parent_id,
	(SELECT COUNT(*) FROM comment_closure_tree 
		 WHERE parent_id = c.id AND depth = 1) as reply_count
			FROM comment c
			JOIN comment_closure_tree ct ON c.id = ct.child_id
			WHERE ct.parent_id = $1 AND ct.depth = 0
		),
	paginated_direct_replies AS (
		SELECT c.id, c.text, ct.depth,
	CASE
		WHEN ct.depth = 0 THEN -1
		ELSE (
			SELECT parent_id FROM comment_closure_tree
			WHERE child_id = c.id and depth = 1
			)
	END as parent_id,
	(SELECT COUNT(*) FROM comment_closure_tree 
		 WHERE parent_id = c.id AND depth = 1) as reply_count
			FROM comment c
				JOIN comment_closure_tree ct on c.id = ct.child_id
			WHERE ct.parent_id = $1 AND ct.depth = 1 AND ct.depth <= $2
			LIMIT CASE WHEN $3 = -1 THEN NULL ELSE $3 END
			OFFSET CASE WHEN $4 = -1 THEN 0 ELSE $4 END
	),
	nested_replies AS (
			SELECT c.id, c.text, ct.depth + 1 as depth,
	CASE
		WHEN ct.depth = 0 THEN -1
		ELSE (
			SELECT parent_id FROM comment_closure_tree
			WHERE child_id = c.id and depth = 1
			)
	END as parent_id,
	(SELECT COUNT(*) FROM comment_closure_tree 
		 WHERE parent_id = c.id AND depth = 1) as reply_count
			FROM comment c
				JOIN comment_closure_tree ct on c.id = ct.child_id
			WHERE ct.parent_id IN (SELECT id from paginated_direct_replies)
			AND ct.depth > 0 AND (ct.depth + 1 <= $2 OR $2 = -1)
	)
	SELECT * FROM root_comment 
	UNION ALL
	SELECT * FROM paginated_direct_replies
	UNION ALL
	SELECT * FROM nested_replies
	ORDER BY depth, id
	`
	offset := 0
	if page != -1 && pageSize != -1 {
		offset = page * pageSize
	}
	rows, err := db.Master.QueryContext(ctx, query, commentID, maxDepth, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []dto.CommentDTO
	for rows.Next() {
		var comment dto.CommentDTO
		if err := rows.Scan(&comment.ID, &comment.Text, &comment.Depth, &comment.ParetID, &comment.AmountOfReplies); err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

func DeleteCommentWithReplies(ctx context.Context, db *database.Database, commentID int64) error {
	query := `
		DELETE FROM comment
		WHERE id IN (
				SELECT child_id FROM comment_closure_tree
				WHERE parent_Id = $1
		)
	`
	_, err := db.Master.ExecContext(ctx, query, commentID)
	return err
}

func SearchInComments(ctx context.Context, db *database.Database, searchText string, page int, pageSize int) ([]dto.CommentDTO, error) {
	query := `
			SELECT id, text FROM comment
			WHERE text_search @@ plainto_tsquery('simple', $1)
			ORDER BY ts_rank(text_search, plainto_tsquery('simple', $1)) DESC
			LIMIT CASE WHEN $2 = -1 THEN NULL ELSE $2 END
			OFFSET CASE WHEN $3 = -1 THEN 0 ELSE $3 END
	`
	offset := 0
	if page != -1 && pageSize != -1 {
		offset = page * pageSize
	}
	rows, err := db.Master.QueryContext(ctx, query, searchText, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []dto.CommentDTO
	for rows.Next() {

		var comment dto.CommentDTO
		if err := rows.Scan(&comment.ID, &comment.Text); err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return comments, nil
}
