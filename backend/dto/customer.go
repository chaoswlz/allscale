package dto

import (
  "time"

  "sarah-project-backend/security"
)

// CustomerDTO represents customer fields read from database.
type CustomerDTO struct {
  ID           int64     `db:"id"`
  Name         string    `db:"name"`
  Email        string    `db:"email"`
  PasswordHash string    `db:"password_hash"`
  CreatedAt    time.Time `db:"created_at"`
  UpdatedAt    time.Time `db:"updated_at"`
}

// VerifyPassword validates a plaintext password against the stored hash.
func (c CustomerDTO) VerifyPassword(plain string) error {
  return security.ComparePassword(c.PasswordHash, plain)
}
