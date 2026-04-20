package repository

import (
	"auth-service/internal/models"
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound          = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

type AuthRepo struct {
	db *pgxpool.Pool
	sq sq.StatementBuilderType
}

func NewAuthRepo(db *pgxpool.Pool) *AuthRepo {
	return &AuthRepo{
		db: db,
		sq: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *AuthRepo) CreateUser(ctx context.Context, user *models.User) (uuid.UUID, error) {
	query, args, err := r.sq.Insert("users").
		Columns("name", "email", "login", "password_hash").
		Values(user.Name, user.Email, user.Login, user.PasswordHash).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return uuid.Nil, fmt.Errorf("AuthRepo.CreateUser: %w", err)
	}
	var userID uuid.UUID
	err = r.db.QueryRow(ctx, query, args...).Scan(&userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return uuid.Nil, ErrUserAlreadyExists
			}
		}
		return uuid.Nil, fmt.Errorf("AuthRepo.CreateUser: %w", err)
	}
	return userID, nil
}
func (r *AuthRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query, args, err := r.sq.
		Select("id", "name", "login", "password_hash", "created_at").
		From("users").
		Where(sq.Eq{"email": email}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("AuthRepo.GetUserByEmail: %w", err)
	}
	var user models.User
	user.Email = email
	err = r.db.QueryRow(ctx, query, args...).Scan(&user.ID, &user.Name, &user.Login, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("AuthRepo.GetUserByEmail: %w", err)
	}
	return &user, nil

}
func (r *AuthRepo) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	query, args, err := r.sq.
		Select("id", "name", "email", "password_hash", "created_at").
		From("users").
		Where(sq.Eq{"login": login}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("AuthRepo.GetUserByLogin: %w", err)
	}
	var user models.User
	user.Login = login
	err = r.db.QueryRow(ctx, query, args...).Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("AuthRepo.GetUserByLogin: %w", err)
	}
	return &user, nil
}
func (r *AuthRepo) GetUsersByIDs(ctx context.Context, ids []string) ([]*models.User, error) {
	query, args, err := r.sq.Select("id", "name", "email", "password_hash", "created_at").
		From("users").
		Where(sq.Eq{"id": ids}).ToSql()
	if err != nil {
		return nil, fmt.Errorf("AuthRepo.GetUsersByIDs: %w", err)
	}
	var users []*models.User
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("AuthRepo.GetUsersByIDs: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var user models.User
		err = rows.Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("AuthRepo.GetUsersByIDs: %w", err)
		}
		users = append(users, &user)
	}
	return users, nil

}

func (r *AuthRepo) SaveTgToken(ctx context.Context, userID string) (string, error) {
	query, args, err := r.sq.
		Insert("tg_links").
		Columns("user_id").
		Values(userID).
		Suffix("RETURNING token").
		ToSql()
	if err != nil {
		return "", fmt.Errorf("AuthRepo.SaveTgToken: %w", err)
	}
	var token string
	err = r.db.QueryRow(ctx, query, args...).Scan(&token)
	if err != nil {
		return "", fmt.Errorf("AuthRepo.SaveTgToken: %w", err)
	}
	return token, nil
}

func (r *AuthRepo) BindTgUser(ctx context.Context, token string, chatID int64) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("AuthRepo.BindTgUser: %w", err)
	}
	defer tx.Rollback(ctx)
	query, args, err := r.sq.
		Delete("tg_links").
		Where(sq.Eq{"token": token}).
		Suffix("RETURNING user_id").
		ToSql()
	if err != nil {
		return fmt.Errorf("AuthRepo.BindTgUser: %w", err)
	}
	var userID uuid.UUID
	err = tx.QueryRow(ctx, query, args...).Scan(&userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("AuthRepo.BindTgUser: %w", err)
	}

	query, args, err = r.sq.Update("users").Set("tg_chat_id", chatID).Where(sq.Eq{"id": userID}).ToSql()
	if err != nil {
		return fmt.Errorf("AuthRepo.BindTgUser: %w", err)
	}
	_, err = tx.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("AuthRepo.BindTgUser: %w", err)
	}
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("AuthRepo.BindTgUser: %w", err)
	}
	return nil
}
