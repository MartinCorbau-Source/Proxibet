package user

import (
	"time"
	"github.com/google/uuid"
)

type status string

const (
	StatusActive     status = "ACTIVE"
	StatusInactive   status = "INACTIVE"
)

type User struct {
	ID          uuid.UUID `json:"id"`
	DisplayName string    `json:"display_name"`
	Email       string    `json:"email"`
	Password    string    `json:"password"`
	Status      status    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}