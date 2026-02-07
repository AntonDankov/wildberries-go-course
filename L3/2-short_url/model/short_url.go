package model

import "time"

type ShortUrl struct {
	ID        int
	Url       string
	CreatedAt time.Time
}
