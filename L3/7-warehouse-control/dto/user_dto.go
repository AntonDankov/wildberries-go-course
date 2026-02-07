package dto

import (
	"wildberries-go-course/L3-7/model"

	"github.com/golang-jwt/jwt/v5"
)

type UserDTO struct {
	Name     string         `json:"name" binding:"required,min=3"`
	Password string         `json:"password" binding:"required,min=6"`
	Role     model.RoleType `json:"role" binding:"required,min=1"`
}

type UserClaimsDTO struct {
	UserID int64          `json:"user_id"`
	Role   model.RoleType `json:"role"`
	jwt.RegisteredClaims
}
