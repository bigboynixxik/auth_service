package service

import (
	"context"
	"testing"

	"auth-service/internal/models"
	"auth-service/internal/repository"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Ручной мок для слоя репозитория
type mockAuthRepo struct {
	createUserFunc     func(ctx context.Context, user *models.User) (uuid.UUID, error)
	getUserByEmailFunc func(ctx context.Context, email string) (*models.User, error)
	getUserByLoginFunc func(ctx context.Context, login string) (*models.User, error)
	getUsersByIDsFunc  func(ctx context.Context, userIDs []string) ([]*models.User, error)
	saveTgTokenFunc    func(ctx context.Context, userID string) (string, error)
	bindTgUserFunc     func(ctx context.Context, token string, chatID int64) error
	getUserChatIDFunc  func(ctx context.Context, userID uuid.UUID) (int64, error)
}

func (m *mockAuthRepo) CreateUser(ctx context.Context, user *models.User) (uuid.UUID, error) {
	if m.createUserFunc != nil {
		return m.createUserFunc(ctx, user)
	}
	return uuid.Nil, nil
}

func (m *mockAuthRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if m.getUserByEmailFunc != nil {
		return m.getUserByEmailFunc(ctx, email)
	}
	return nil, nil
}

func (m *mockAuthRepo) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	if m.getUserByLoginFunc != nil {
		return m.getUserByLoginFunc(ctx, login)
	}
	return nil, nil
}

func (m *mockAuthRepo) GetUsersByIDs(ctx context.Context, userIDs []string) ([]*models.User, error) {
	if m.getUsersByIDsFunc != nil {
		return m.getUsersByIDsFunc(ctx, userIDs)
	}
	return nil, nil
}

func (m *mockAuthRepo) SaveTgToken(ctx context.Context, userID string) (string, error) {
	if m.saveTgTokenFunc != nil {
		return m.saveTgTokenFunc(ctx, userID)
	}
	return "", nil
}

func (m *mockAuthRepo) BindTgUser(ctx context.Context, token string, chatID int64) error {
	if m.bindTgUserFunc != nil {
		return m.bindTgUserFunc(ctx, token, chatID)
	}
	return nil
}

func (m *mockAuthRepo) GetUserChatID(ctx context.Context, userID uuid.UUID) (int64, error) {
	if m.getUserChatIDFunc != nil {
		return m.getUserChatIDFunc(ctx, userID)
	}
	return 0, nil
}

func TestAuthService_Register(t *testing.T) {
	secret := []byte("super_secret_key")
	ctx := context.Background()

	tests := []struct {
		name      string
		mockRepo  *mockAuthRepo
		email     string
		login     string
		password  string
		wantError bool
	}{
		{
			name: "Успешная регистрация",
			mockRepo: &mockAuthRepo{
				createUserFunc: func(ctx context.Context, user *models.User) (uuid.UUID, error) {
					return uuid.New(), nil
				},
			},
			email:     "test@mail.com",
			login:     "test",
			password:  "12345",
			wantError: false,
		},
		{
			name: "Пользователь уже существует",
			mockRepo: &mockAuthRepo{
				createUserFunc: func(ctx context.Context, user *models.User) (uuid.UUID, error) {
					return uuid.Nil, repository.ErrUserAlreadyExists
				},
			},
			email:     "exist@mail.com",
			login:     "exist",
			password:  "12345",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewAuthService(tt.mockRepo, secret)
			token, err := svc.Register(ctx, tt.email, tt.login, "Test Name", tt.password)

			if (err != nil) != tt.wantError {
				t.Errorf("Register() error = %v, wantError %v", err, tt.wantError)
			}
			if !tt.wantError && token == "" {
				t.Errorf("Register() expected valid token, got empty")
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	secret := []byte("super_secret_key")
	ctx := context.Background()

	// Генерируем реальный хэш для тестов
	validHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	tests := []struct {
		name      string
		mockRepo  *mockAuthRepo
		email     string
		password  string
		wantError bool
	}{
		{
			name: "Успешный логин",
			mockRepo: &mockAuthRepo{
				getUserByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
					return &models.User{
						ID:           uuid.New(),
						Email:        "test@mail.com",
						PasswordHash: string(validHash),
					}, nil
				},
			},
			email:     "test@mail.com",
			password:  "password123",
			wantError: false,
		},
		{
			name: "Неверный пароль",
			mockRepo: &mockAuthRepo{
				getUserByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
					return &models.User{
						ID:           uuid.New(),
						Email:        "test@mail.com",
						PasswordHash: string(validHash),
					}, nil
				},
			},
			email:     "test@mail.com",
			password:  "wrong_password",
			wantError: true,
		},
		{
			name: "Пользователь не найден",
			mockRepo: &mockAuthRepo{
				getUserByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
					return nil, repository.ErrNotFound
				},
			},
			email:     "notfound@mail.com",
			password:  "123",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewAuthService(tt.mockRepo, secret)
			token, err := svc.Login(ctx, tt.email, tt.password)

			if (err != nil) != tt.wantError {
				t.Errorf("Login() error = %v, wantError %v", err, tt.wantError)
			}
			if !tt.wantError && token == "" {
				t.Errorf("Login() expected valid token, got empty")
			}
		})
	}
}
