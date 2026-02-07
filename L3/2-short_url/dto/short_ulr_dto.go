package dto

type ShortUrlDTO struct {
	Url string `json:"url" binding:requered`
}
