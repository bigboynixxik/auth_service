package transport

import (
	"auth-service/internal/models"
	"context"

	"github.com/google/uuid"
)

type AuthService interface {
	Register(ctx context.Context, email, login, name, password string) (string, error)
	Login(ctx context.Context, email, password string) (string, error)
	GetUserInfoByLogin(ctx context.Context, login string) (*models.UserInfo, error)
	GetUsersInfo(ctx context.Context, userIDs []string) (map[string]*models.UserInfo, error)
	SaveTgToken(ctx context.Context, userID string) (string, error)
	BindTgUser(ctx context.Context, token string, chatID int64) error
	GetUserChatID(ctx context.Context, userID uuid.UUID) (int64, error)
}
