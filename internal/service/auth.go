package service

import (
	"auth-service/internal/models"
	"auth-service/internal/repository"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New("invalid email or password")

type AuthService struct {
	repo      AuthRepository
	jwtSecret []byte
}

func NewAuthService(repo AuthRepository, jwtSecret []byte) *AuthService {
	return &AuthService{repo: repo, jwtSecret: jwtSecret}
}

func (as *AuthService) Register(ctx context.Context, email, login, name, password string) (string, error) {
	// TODO: Добавить проверку на длину пароля и его валидность
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("service.Register failed to generate password hash: %w", err)
	}
	// TODO: Добавить проверку на валидность email
	user := &models.User{
		Email:        email,
		Login:        login,
		Name:         name,
		PasswordHash: string(passwordHash),
	}
	id, err := as.repo.CreateUser(ctx, user)
	if err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			return "", repository.ErrUserAlreadyExists
		}
		return "", fmt.Errorf("service.Register failed to create user: %w", err)
	}
	tokenString, err := as.generateToken(id)
	if err != nil {
		return "", fmt.Errorf("service.Register failed to generate token: %w", err)
	}
	return tokenString, nil
}
func (as *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := as.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", repository.ErrNotFound
		}
		return "", fmt.Errorf("service.Login failed to get user by email: %w", err)
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", ErrInvalidCredentials
	}
	tokenString, err := as.generateToken(user.ID)
	if err != nil {
		return "", fmt.Errorf("service.Login failed to generate token: %w", err)
	}
	return tokenString, nil
}

func (as *AuthService) GetUserInfoByLogin(ctx context.Context, login string) (*models.UserInfo, error) {
	user, err := as.repo.GetUserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("service.GetUserInfoByLogin failed to get user by login: %w", err)
	}
	var userInfo models.UserInfo
	userInfo.Login = login
	userInfo.ID = user.ID
	userInfo.Name = user.Name
	return &userInfo, nil
}

func (as *AuthService) GetUsersInfo(ctx context.Context, userIDs []string) (map[string]*models.UserInfo, error) {
	users, err := as.repo.GetUsersByIDs(ctx, userIDs)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("service.GetUsersInfo failed to get users info: %w", err)
	}
	usersInfoMap := make(map[string]*models.UserInfo)
	for _, user := range users {
		userInfo := models.UserInfo{
			ID:    user.ID,
			Name:  user.Name,
			Login: user.Login,
		}
		usersInfoMap[userInfo.ID.String()] = &userInfo
	}
	return usersInfoMap, nil
}

func (as *AuthService) generateToken(userID uuid.UUID) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userID.String()
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()
	tokenString, err := token.SignedString(as.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("service.generateToken failed to sign token: %w", err)
	}
	return tokenString, nil
}
