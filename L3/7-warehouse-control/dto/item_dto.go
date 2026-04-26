package dto

type ItemDTO struct {
	ID      int64   `json:"id"`
	OwnerID int64   `json:"owner_id"`
	Name    string  `json:"name" binding:"required,min=1"`
	Price   float64 `json:"price" binding:"required,min=0"`
	Amount  int     `json:"amount" binding:"required,min=0"`
}
