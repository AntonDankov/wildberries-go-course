package dto

type RecordDTO struct {
	ID       int64   `json:"id"`
	Type     int     `json:"type" binding:"required`
	Category int     `json:"category" binding:"required"`
	Amount   float64 `json:"amount" binding:"required,min=0"`
	Date     string  `json:"date" binding:"required"`
}
