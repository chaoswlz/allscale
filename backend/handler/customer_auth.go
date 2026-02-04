package handler

import (
  "context"
  "database/sql"
  "fmt"
  "net/http"
  "strings"
  "time"
)

func authenticateCustomer(r *http.Request, db *sql.DB) (string, error) {
  apiKey := strings.TrimSpace(r.Header.Get("X-API-Key"))
  if apiKey == "" {
    return "", fmt.Errorf("missing api key")
  }

  merchantName := strings.TrimSpace(r.Header.Get("X-Merchant-Name"))
  if merchantName == "" {
    return "", fmt.Errorf("missing merchant name")
  }

  ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
  defer cancel()

  var matched string
  err := db.QueryRowContext(ctx, `
    SELECT merchant_name
    FROM customer_api_keys
    WHERE api_key = ? AND merchant_name = ? AND active = 1
    LIMIT 1
  `, apiKey, merchantName).Scan(&matched)
  if err != nil {
    if err == sql.ErrNoRows {
      return "", fmt.Errorf("invalid api key")
    }
    return "", err
  }

  return matched, nil
}
