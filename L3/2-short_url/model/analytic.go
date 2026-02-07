package model

import "time"

type Analytic struct {
	ID        int64
	UserAgent string
	VisitTime time.Time
	URLID     int64
}
