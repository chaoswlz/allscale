package dto

import (
  "time"

  "sarah-project-backend/security"
)

// AdminDTO represents admin fields read from database.
type AdminDTO struct {
  ID           int64     `db:"id"`
  Username     string    `db:"username"`
  Email        string    `db:"email"`
  PasswordHash string    `db:"password_hash"`
  CreatedAt    time.Time `db:"created_at"`
  UpdatedAt    time.Time `db:"updated_at"`
}

// VerifyPassword validates a plaintext password against the stored hash.
func (a AdminDTO) VerifyPassword(plain string) error {
  return security.ComparePassword(a.PasswordHash, plain)
}
