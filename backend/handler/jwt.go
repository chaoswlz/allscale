package handler

import (
  "fmt"
  "net/http"
  "strings"
  "time"

  "github.com/golang-jwt/jwt/v5"
)

type jwtClaims struct {
  Role string `json:"role"`
  jwt.RegisteredClaims
}

func createToken(userID int64, role string, secret string, issuer string, ttl time.Duration) (string, error) {
  now := time.Now()
  claims := jwtClaims{
    Role: role,
    RegisteredClaims: jwt.RegisteredClaims{
      Subject:   fmt.Sprintf("%d", userID),
      Issuer:    issuer,
      IssuedAt:  jwt.NewNumericDate(now),
      ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
    },
  }

  token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
  return token.SignedString([]byte(secret))
}

func authenticateRequest(r *http.Request, secret string) (*jwtClaims, error) {
  authHeader := r.Header.Get("Authorization")
  if authHeader == "" {
    return nil, fmt.Errorf("missing authorization header")
  }

  parts := strings.SplitN(authHeader, " ", 2)
  if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
    return nil, fmt.Errorf("invalid authorization header")
  }

  tokenStr := strings.TrimSpace(parts[1])
  if tokenStr == "" {
    return nil, fmt.Errorf("empty token")
  }

  token, err := jwt.ParseWithClaims(tokenStr, &jwtClaims{}, func(token *jwt.Token) (any, error) {
    if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
      return nil, fmt.Errorf("unexpected signing method")
    }
    return []byte(secret), nil
  })
  if err != nil || !token.Valid {
    return nil, fmt.Errorf("invalid token")
  }

  claims, ok := token.Claims.(*jwtClaims)
  if !ok {
    return nil, fmt.Errorf("invalid claims")
  }
  return claims, nil
}
