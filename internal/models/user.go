package models

import (
	"time"

	"github.com/google/uuid"
)

type UserInfo struct {
	ID    uuid.UUID
	Name  string
	Login string
}

type User struct {
	ID           uuid.UUID
	Name         string
	Email        string
	Login        string
	PasswordHash string
	CreatedAt    time.Time
}
