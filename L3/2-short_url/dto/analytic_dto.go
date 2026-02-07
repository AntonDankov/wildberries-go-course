package dto

import "time"

type Analytic struct {
	ID        int64     `json:"id"`
	UserAgent string    `json:"user_agent"`
	VisitTime time.Time `json:"visit_time"`
	URLID     int64     `json:"url_id"`
}

type AnalyticAggregatedByDate struct {
	Date       time.Time `json: "date"`
	VisitCount int64     `json: "visit_count"`
}

type AnalyticAggregatedByUserAgent struct {
	UserAgent  string `json:"user_agent"`
	VisitCount int64  `json:"visit_count"`
}
