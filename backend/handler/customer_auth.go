package handler

import (
  "fmt"
  "net/http"
  "strings"
)

type CustomerAuthConfig struct {
  APIKey       string
  MerchantName string
}

func authenticateCustomer(r *http.Request, cfg CustomerAuthConfig) (string, error) {
  apiKey := strings.TrimSpace(r.Header.Get("X-API-Key"))
  if apiKey == "" {
    return "", fmt.Errorf("missing api key")
  }
  if apiKey != cfg.APIKey {
    return "", fmt.Errorf("invalid api key")
  }

  merchantName := strings.TrimSpace(r.Header.Get("X-Merchant-Name"))
  if merchantName == "" {
    return "", fmt.Errorf("missing merchant name")
  }
  if merchantName != cfg.MerchantName {
    return "", fmt.Errorf("invalid merchant name")
  }

  return merchantName, nil
}
