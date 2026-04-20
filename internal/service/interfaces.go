package service

import (
	"auth-service/internal/models"
	"context"

	"github.com/google/uuid"
)

type AuthRepository interface {
	CreateUser(ctx context.Context, user *models.User) (uuid.UUID, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)
	GetUsersByIDs(ctx context.Context, ids []string) ([]*models.User, error)
	SaveTgToken(ctx context.Context, userID string) (string, error)
	BindTgUser(ctx context.Context, token string, chatID int64) error
}
