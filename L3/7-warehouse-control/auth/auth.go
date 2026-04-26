package auth

import (
	"context"
	"fmt"
	"strings"
	"wildberries-go-course/L3-7/database"
	"wildberries-go-course/L3-7/dto"
	"wildberries-go-course/L3-7/model"
	"wildberries-go-course/L3-7/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
)

var JwtSecret = []byte("very-secret-key")

func Authorize(ctx context.Context, db database.DBTX, c *ginext.Context, allowedRoles model.RoleType) (dto.UserClaimsDTO, error) {
	authHeader := c.GetHeader("Authorization")
	zlog.Logger.Debug().Msgf("auth header: %v", authHeader)
	userClaims, err := GetUserClaims(authHeader)
	if err != nil {
		return userClaims, err
	}
	isRoleAccepted := CheckRequiredRoleByUserClaims(userClaims, allowedRoles)
	if !isRoleAccepted {
		return userClaims, fmt.Errorf("no access for user role")
	}
	return userClaims, nil
}

func GetUserClaims(authHeader string) (dto.UserClaimsDTO, error) {
	var userClaims dto.UserClaimsDTO
	if authHeader == "" {
		return userClaims, fmt.Errorf("authorization header required")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if len(tokenString) == len(authHeader) {
		return userClaims, fmt.Errorf("invalid authorization format")
	}
	zlog.Logger.Debug().Msgf("auth token: %v", tokenString)
	token, err := jwt.ParseWithClaims(tokenString, &userClaims, func(token *jwt.Token) (any, error) {
		return JwtSecret, nil
	})

	if err != nil || !token.Valid {
		return userClaims, fmt.Errorf("invalid token: %w", err)
	}

	return userClaims, nil
}

// I'm more preferring check important things like role through database,
// because otherwise if account get stolen we can't stop it for now
// then we should introduce some mechanisms to block account and it still might be another database request
func CheckRequiredRoleByDatabase(ctx context.Context, db database.DBTX, userID int64, allowedRoles model.RoleType) bool {
	user, err := repository.GetUserByID(ctx, db, userID)
	if err != nil {
		zlog.Logger.Fatal().Msgf("failed to get user role: %v", err)
		return false
	}

	return (user.Role & allowedRoles) != 0
}

func CheckRequiredRoleByUserClaims(userClaims dto.UserClaimsDTO, allowedRoles model.RoleType) bool {
	return (userClaims.Role & allowedRoles) != 0
}
