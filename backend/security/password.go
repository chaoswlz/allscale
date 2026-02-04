package security

import "golang.org/x/crypto/bcrypt"

// HashPassword hashes the plaintext password using bcrypt.
func HashPassword(plain string) (string, error) {
  hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
  if err != nil {
    return "", err
  }
  return string(hash), nil
}

// ComparePassword checks a plaintext password against a bcrypt hash.
func ComparePassword(hash, plain string) error {
  return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
}
