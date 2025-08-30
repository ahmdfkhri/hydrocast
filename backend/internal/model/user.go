package model

import (
	"time"

	"github.com/ahmdfkhri/hydrocast/backend/internal/types"
	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID      `db:"id"`
	Username  string         `db:"username"`
	Email     string         `db:"email"`
	Password  string         `db:"password"`
	Role      types.UserRole `db:"role"`
	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`
}
