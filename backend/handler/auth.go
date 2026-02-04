package handler

import (
  "context"
  "database/sql"
  "encoding/json"
  "log"
  "net/http"
  "time"

  "sarah-project-backend/dto"
)

type adminLoginRequest struct {
  Username string `json:"username"`
  Password string `json:"password"`
}


type adminLoginResponse struct {
  ID       int64  `json:"id"`
  Username string `json:"username"`
  Email    string `json:"email"`
  Token    string `json:"token"`
}


type AuthConfig struct {
  JWTSecret string
  JWTIssuer string
  JWTTTL    time.Duration
}

// AdminLogin handles admin login requests.
func AdminLogin(db *sql.DB, cfg AuthConfig) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
      writeError(w, http.StatusMethodNotAllowed, "method not allowed")
      return
    }

    var req adminLoginRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
      writeError(w, http.StatusBadRequest, "invalid json body")
      return
    }
    if req.Username == "" || req.Password == "" {
      writeError(w, http.StatusBadRequest, "username and password required")
      return
    }

    admin, err := getAdminByUsername(r.Context(), db, req.Username)
    if err != nil {
      if err == sql.ErrNoRows {
        writeError(w, http.StatusUnauthorized, "invalid credentials")
        return
      }
      log.Printf("admin login query error: %v", err)
      writeError(w, http.StatusInternalServerError, "server error")
      return
    }
    if err := admin.VerifyPassword(req.Password); err != nil {
      writeError(w, http.StatusUnauthorized, "invalid credentials")
      return
    }

    token, err := createToken(admin.ID, "admin", cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTTTL)
    if err != nil {
      log.Printf("admin login token error: %v", err)
      writeError(w, http.StatusInternalServerError, "server error")
      return
    }

    resp := adminLoginResponse{
      ID:       admin.ID,
      Username: admin.Username,
      Email:    admin.Email,
      Token:    token,
    }
    writeJSON(w, http.StatusOK, resp)
  }
}

func getAdminByUsername(ctx context.Context, db *sql.DB, username string) (dto.AdminDTO, error) {
  ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
  defer cancel()

  var admin dto.AdminDTO
  row := db.QueryRowContext(ctx, `
    SELECT id, username, email, password_hash, created_at, updated_at
    FROM admin_users
    WHERE username = ?
    LIMIT 1
  `, username)
  if err := row.Scan(
    &admin.ID,
    &admin.Username,
    &admin.Email,
    &admin.PasswordHash,
    &admin.CreatedAt,
    &admin.UpdatedAt,
  ); err != nil {
    return dto.AdminDTO{}, err
  }
  return admin, nil
}

