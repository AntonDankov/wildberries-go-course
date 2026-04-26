package repository

import (
	"context"
	"fmt"
	"strings"
	"wildberries-go-course/L3-6/database"
	"wildberries-go-course/L3-6/dto"
)

func GetAnalytics(ctx context.Context, db database.DBTX, filter *dto.AnalyticsFilter) (*dto.Analytics, error) {
	baseQuery := `
		SELECT 
			COALESCE(SUM(amount), 0) as sum,
			COALESCE(AVG(amount), 0) as average,
			COUNT(*) as count,
			COALESCE(PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY amount), 0) as median,
			COALESCE(PERCENTILE_CONT(0.9) WITHIN GROUP (ORDER BY amount), 0) as percentile_90
		FROM record
	`

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

	query := baseQuery
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	var analytics dto.Analytics
	err := db.QueryRowContext(ctx, query, args...).Scan(
		&analytics.Sum,
		&analytics.Average,
		&analytics.Count,
		&analytics.Median,
		&analytics.Percentile,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get analytics: %v", err)
	}

	return &analytics, nil
}
