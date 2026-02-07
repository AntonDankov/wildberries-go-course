package repository

import (
	"context"
	"fmt"
	"time"
	database "wildberries-go-course/L3-2/database"
	dto "wildberries-go-course/L3-2/dto"
	model "wildberries-go-course/L3-2/model"
)

func GetAnalyticsFull(ctx context.Context, db *database.Database, urlID int64) ([]dto.Analytic, error) {
	query := `SELECT id, user_agent,visit_time, short_url_id FROM analytics
	WHERE short_url_id = $1
	`
	var urlAnalytics []dto.Analytic
	rows, err := db.Master.QueryContext(ctx, query,
		urlID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add ShortUrl: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var analytic dto.Analytic
		if err := rows.Scan(&analytic.ID, &analytic.UserAgent, &analytic.VisitTime, &analytic.URLID); err != nil {
			return nil, err
		}
		urlAnalytics = append(urlAnalytics, analytic)
	}
	return urlAnalytics, nil
}

func AddAnalytic(ctx context.Context, db *database.Database, analytic model.Analytic) (int64, error) {
	query := `
		INSERT INTO analytics (user_agent, visit_time, short_url_id)
	VALUES ($1,$2,$3)
	RETURNING id
	`
	var id int64
	err := db.Master.QueryRowContext(ctx, query,
		analytic.UserAgent, analytic.VisitTime, analytic.URLID,
	).Scan(&id)
	if err != nil {
		return -1, nil
	}
	return id, nil
}

func GetAnalyticsAggreatedByDay(
	ctx context.Context,
	db *database.Database,
	urlID int64,
	startDate, endDate time.Time,
) ([]dto.AnalyticAggregatedByDate, error) {
	query := `
		SELECT DATE_TRUNC('day',visit_time) AS day, COUNT(*) as count
		FROM analytics
		WHERE short_url_id = $1 AND visit_time >= $2 AND visit_time < $3
		GROUP BY day
		ORDER by day
	`
	rows, err := db.Master.QueryContext(ctx, query, urlID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var urlAnalytics []dto.AnalyticAggregatedByDate
	for rows.Next() {
		var analytic dto.AnalyticAggregatedByDate
		if err := rows.Scan(&analytic.Date, &analytic.VisitCount); err != nil {
			return nil, err
		}
		urlAnalytics = append(urlAnalytics, analytic)
	}

	return urlAnalytics, nil
}

func GetAnalyticsAggregatedByMonth(
	ctx context.Context,
	db *database.Database,
	urlID int64,
	startDate, endDate time.Time,
) ([]dto.AnalyticAggregatedByDate, error) {
	query := `
        SELECT DATE_TRUNC('month', visit_time) AS month, COUNT(*) AS count
        FROM analytics
        WHERE short_url_id = $1
          AND visit_time >= $2
          AND visit_time <  $3
        GROUP BY month
        ORDER BY month
    `
	rows, err := db.Master.QueryContext(ctx, query, urlID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var urlAnalytics []dto.AnalyticAggregatedByDate
	for rows.Next() {
		var analytic dto.AnalyticAggregatedByDate
		if err := rows.Scan(&analytic.Date, &analytic.VisitCount); err != nil {
			return nil, err
		}
		urlAnalytics = append(urlAnalytics, analytic)
	}

	return urlAnalytics, nil
}

func GetAnalyticsAggregatedByUserAgent(
	ctx context.Context,
	db *database.Database,
	urlID int64,
	startDate, endDate time.Time,
) ([]dto.AnalyticAggregatedByUserAgent, error) {
	query := `
        SELECT user_agent, COUNT(*) AS count
        FROM analytics
        WHERE short_url_id = $1
          AND visit_time >= $2
          AND visit_time <  $3
        GROUP BY user_agent
        ORDER BY count DESC
    `
	rows, err := db.Master.QueryContext(ctx, query, urlID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []dto.AnalyticAggregatedByUserAgent
	for rows.Next() {
		var analytic dto.AnalyticAggregatedByUserAgent
		if err := rows.Scan(&analytic.UserAgent, &analytic.VisitCount); err != nil {
			return nil, err
		}
		result = append(result, analytic)
	}

	return result, nil
}
