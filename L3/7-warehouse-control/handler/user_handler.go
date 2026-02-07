package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"
	"wildberries-go-course/L3-7/auth"
	"wildberries-go-course/L3-7/database"
	"wildberries-go-course/L3-7/dto"
	"wildberries-go-course/L3-7/model"
	"wildberries-go-course/L3-7/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/wb-go/wbf/ginext"
	"golang.org/x/crypto/bcrypt"
)

func Register(ctx context.Context, db *database.Database) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		var req dto.UserDTO

		if err := c.ShouldBindJSON(&req); err != nil {
			addJSONWithError(c, http.StatusBadRequest, fmt.Errorf("invalid request: %w", err))
			return
		}

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			addJSONWithError(c, http.StatusInternalServerError, fmt.Errorf("failed to hash password: %w", err))
			return
		}

		userID, err := repository.CreateUser(ctx, db.Master, req.Name, string(passwordHash), req.Role)
		if err != nil {
			addJSONWithError(c, http.StatusInternalServerError, fmt.Errorf("failed to create user: %w", err))
			return
		}

		c.JSON(http.StatusCreated, ginext.H{
			"user_id": userID,
			"name":    req.Name,
			"role":    req.Role,
		})
	}
}

func Login(ctx context.Context, db *database.Database) ginext.HandlerFunc {
	return func(c *ginext.Context) {
		var req dto.UserDTO

		if err := c.ShouldBindJSON(&req); err != nil {
			addJSONWithError(c, http.StatusBadRequest, fmt.Errorf("invalid request: %w", err))
			return
		}

		user, err := repository.GetUserByName(ctx, db.Master, req.Name)
		if err != nil {
			addJSONWithError(c, http.StatusUnauthorized, fmt.Errorf("invalid credentials"))
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
		if err != nil {
			addJSONWithError(c, http.StatusUnauthorized, fmt.Errorf("invalid credentials"))
			return
		}

		token, err := generateToken(user.ID, user.Role)
		if err != nil {
			addJSONWithError(c, http.StatusInternalServerError, fmt.Errorf("failed to generate token: %w", err))
			return
		}

		c.JSON(http.StatusOK, ginext.H{
			"token":   token,
			"user_id": user.ID,
			"name":    user.Name,
			"role":    user.Role,
		})
	}
}

func generateToken(userID int64, role model.RoleType) (string, error) {
	claims := dto.UserClaimsDTO{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(auth.JwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %v", err)
	}

	return tokenString, nil
}

func addJSONWithError(c *ginext.Context, httpCode int, err error) {
	c.JSON(httpCode, ginext.H{
		"error": err.Error(),
	})
}
