package transport

import (
	"auth-service/internal/models"
	"context"
)

type AuthService interface {
	Register(ctx context.Context, email, login, name, password string) (string, error)
	Login(ctx context.Context, email, password string) (string, error)
	GetUserInfoByLogin(ctx context.Context, login string) (*models.UserInfo, error)
	GetUsersInfo(ctx context.Context, userIDs []string) (map[string]*models.UserInfo, error)
}
