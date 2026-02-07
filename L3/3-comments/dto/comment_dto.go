package dto

type CreateCommentDTO struct {
	Text     string `json:"text" binding:"required"`
	ParentID int64  `json:"parent_id"`
}

type CommentDTO struct {
	ID              int64  `json:"id"`
	Text            string `json:"text"`
	ParetID         int64  `json:"parent_id"`
	Depth           int    `json:"depth,omitempty"`
	AmountOfReplies int    `json:"amount_of_replies"`
	// CreatedAt time.Time `json:"created_at,omitempty"`
}
