package handler

import (
  "context"
  "database/sql"
  "encoding/json"
  "log"
  "net/http"
  "strconv"
  "strings"
  "time"
)

type createOrderRequest struct {
  TransactionNetwork string   `json:"transaction_network"`
  TransactionAsset   string   `json:"transaction_asset"`
  TXID               string   `json:"txid"`
  Amount             *float64 `json:"amount"`
  BeneficiaryName    string   `json:"beneficiary_name"`
  BankCountry        string   `json:"bank_country"`
  BankName           string   `json:"bank_name"`
  IBAN               string   `json:"iban"`
  SWIFT              string   `json:"swift"`
  ReferenceNote      *string  `json:"reference_note"`
}

type createOrderResponse struct {
  ID int64 `json:"id"`
}

type listOrdersResponse struct {
  Total    int64           `json:"total"`
  Page     int             `json:"page"`
  PageSize int             `json:"page_size"`
  Orders   []orderResponse `json:"orders"`
}

type orderDetailResponse struct {
  Order orderResponse `json:"order"`
}

type orderResponse struct {
  ID                 int64    `json:"id"`
  TransactionNetwork string   `json:"transaction_network"`
  TransactionAsset   string   `json:"transaction_asset"`
  TXID               string   `json:"txid"`
  Amount             *float64 `json:"amount"`
  BeneficiaryName    string   `json:"beneficiary_name"`
  BankCountry        string   `json:"bank_country"`
  BankName           string   `json:"bank_name"`
  IBAN               string   `json:"iban"`
  SWIFT              string   `json:"swift"`
  ReferenceNote      *string  `json:"reference_note"`
  Status             string   `json:"status"`
  CreatedAt          time.Time `json:"created_at"`
}

// CreateOrder allows a customer to create a new order.
func CreateOrder(db *sql.DB, cfg CustomerAuthConfig) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
      writeError(w, http.StatusMethodNotAllowed, "method not allowed")
      return
    }

    merchantName, err := authenticateCustomer(r, cfg)
    if err != nil {
      writeError(w, http.StatusUnauthorized, "unauthorized")
      return
    }

    var req createOrderRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
      writeError(w, http.StatusBadRequest, "invalid json body")
      return
    }

    if err := validateCreateOrder(req); err != nil {
      writeError(w, http.StatusBadRequest, err.Error())
      return
    }

    id, err := insertOrder(r.Context(), db, merchantName, req)
    if err != nil {
      log.Printf("create order error: %v", err)
      writeError(w, http.StatusInternalServerError, "server error")
      return
    }

    writeJSON(w, http.StatusCreated, createOrderResponse{ID: id})
  }
}

// ListCustomerOrders allows a customer to list their orders with pagination.
func ListCustomerOrders(db *sql.DB, cfg CustomerAuthConfig) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
      writeError(w, http.StatusMethodNotAllowed, "method not allowed")
      return
    }

    merchantName, err := authenticateCustomer(r, cfg)
    if err != nil {
      writeError(w, http.StatusUnauthorized, "unauthorized")
      return
    }

    page, pageSize, err := parsePagination(r)
    if err != nil {
      writeError(w, http.StatusBadRequest, err.Error())
      return
    }

    total, orders, err := listOrdersByMerchant(r.Context(), db, merchantName, page, pageSize)
    if err != nil {
      log.Printf("list orders error: %v", err)
      writeError(w, http.StatusInternalServerError, "server error")
      return
    }

    writeJSON(w, http.StatusOK, listOrdersResponse{
      Total:    total,
      Page:     page,
      PageSize: pageSize,
      Orders:   orders,
    })
  }
}

// GetCustomerOrder returns a single order by ID for the authenticated merchant.
func GetCustomerOrder(db *sql.DB, cfg CustomerAuthConfig) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
      writeError(w, http.StatusMethodNotAllowed, "method not allowed")
      return
    }

    merchantName, err := authenticateCustomer(r, cfg)
    if err != nil {
      writeError(w, http.StatusUnauthorized, "unauthorized")
      return
    }

    rawID := r.URL.Query().Get("id")
    if rawID == "" {
      writeError(w, http.StatusBadRequest, "id is required")
      return
    }
    orderID, err := strconv.ParseInt(rawID, 10, 64)
    if err != nil || orderID <= 0 {
      writeError(w, http.StatusBadRequest, "invalid id")
      return
    }

    order, err := getOrderByID(r.Context(), db, merchantName, orderID)
    if err != nil {
      if err == sql.ErrNoRows {
        writeError(w, http.StatusNotFound, "order not found")
        return
      }
      log.Printf("get order error: %v", err)
      writeError(w, http.StatusInternalServerError, "server error")
      return
    }

    writeJSON(w, http.StatusOK, orderDetailResponse{Order: order})
  }
}

func validateCreateOrder(req createOrderRequest) error {
  if req.TransactionNetwork == "" ||
    req.TransactionAsset == "" ||
    req.TXID == "" ||
    req.BeneficiaryName == "" ||
    req.BankCountry == "" ||
    req.BankName == "" ||
    req.IBAN == "" ||
    req.SWIFT == "" {
    return errBadRequest("missing required fields")
  }

  if !isAllowed(req.TransactionNetwork, []string{"TRON", "BSC", "Ethereum"}) {
    return errBadRequest("invalid transaction_network")
  }
  if !isAllowed(req.TransactionAsset, []string{"USDT", "USDC"}) {
    return errBadRequest("invalid transaction_asset")
  }
  if !isAllowed(req.BankCountry, []string{"Canada", "United States"}) {
    return errBadRequest("invalid bank_country")
  }

  if req.Amount != nil && *req.Amount < 0 {
    return errBadRequest("amount must be non-negative")
  }

  return nil
}

func insertOrder(ctx context.Context, db *sql.DB, merchantName string, req createOrderRequest) (int64, error) {
  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()

  var amount sql.NullFloat64
  if req.Amount != nil {
    amount = sql.NullFloat64{Float64: *req.Amount, Valid: true}
  }

  var note sql.NullString
  if req.ReferenceNote != nil && strings.TrimSpace(*req.ReferenceNote) != "" {
    note = sql.NullString{String: *req.ReferenceNote, Valid: true}
  }

  result, err := db.ExecContext(ctx, `
    INSERT INTO orders (
      merchant_name,
      transaction_network,
      transaction_asset,
      txid,
      amount,
      beneficiary_name,
      bank_country,
      bank_name,
      iban,
      swift,
      reference_note,
      status
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
  `,
    merchantName,
    req.TransactionNetwork,
    req.TransactionAsset,
    req.TXID,
    amount,
    req.BeneficiaryName,
    req.BankCountry,
    req.BankName,
    req.IBAN,
    req.SWIFT,
    note,
    "Processing",
  )
  if err != nil {
    return 0, err
  }

  return result.LastInsertId()
}

func listOrdersByMerchant(ctx context.Context, db *sql.DB, merchantName string, page int, pageSize int) (int64, []orderResponse, error) {
  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()

  var total int64
  if err := db.QueryRowContext(ctx, `
    SELECT COUNT(*)
    FROM orders
    WHERE merchant_name = ?
  `, merchantName).Scan(&total); err != nil {
    return 0, nil, err
  }

  offset := (page - 1) * pageSize
  rows, err := db.QueryContext(ctx, `
    SELECT id, transaction_network, transaction_asset, txid, amount, beneficiary_name,
           bank_country, bank_name, iban, swift, reference_note, status, created_at
    FROM orders
    WHERE merchant_name = ?
    ORDER BY id DESC
    LIMIT ? OFFSET ?
  `, merchantName, pageSize, offset)
  if err != nil {
    return 0, nil, err
  }
  defer rows.Close()

  orders := make([]orderResponse, 0)
  for rows.Next() {
    var (
      amount sql.NullFloat64
      note   sql.NullString
      order  orderResponse
    )
    if err := rows.Scan(
      &order.ID,
      &order.TransactionNetwork,
      &order.TransactionAsset,
      &order.TXID,
      &amount,
      &order.BeneficiaryName,
      &order.BankCountry,
      &order.BankName,
      &order.IBAN,
      &order.SWIFT,
      &note,
      &order.Status,
      &order.CreatedAt,
    ); err != nil {
      return 0, nil, err
    }
    if amount.Valid {
      order.Amount = &amount.Float64
    }
    if note.Valid && strings.TrimSpace(note.String) != "" {
      order.ReferenceNote = &note.String
    }
    orders = append(orders, order)
  }
  if err := rows.Err(); err != nil {
    return 0, nil, err
  }

  return total, orders, nil
}

func getOrderByID(ctx context.Context, db *sql.DB, merchantName string, orderID int64) (orderResponse, error) {
  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()

  var (
    amount sql.NullFloat64
    note   sql.NullString
    order  orderResponse
  )

  row := db.QueryRowContext(ctx, `
    SELECT id, transaction_network, transaction_asset, txid, amount, beneficiary_name,
           bank_country, bank_name, iban, swift, reference_note, status, created_at
    FROM orders
    WHERE merchant_name = ? AND id = ?
    LIMIT 1
  `, merchantName, orderID)
  if err := row.Scan(
    &order.ID,
    &order.TransactionNetwork,
    &order.TransactionAsset,
    &order.TXID,
    &amount,
    &order.BeneficiaryName,
    &order.BankCountry,
    &order.BankName,
    &order.IBAN,
    &order.SWIFT,
    &note,
    &order.Status,
    &order.CreatedAt,
  ); err != nil {
    return orderResponse{}, err
  }

  if amount.Valid {
    order.Amount = &amount.Float64
  }
  if note.Valid && strings.TrimSpace(note.String) != "" {
    order.ReferenceNote = &note.String
  }

  return order, nil
}

func isAllowed(value string, allowed []string) bool {
  for _, item := range allowed {
    if value == item {
      return true
    }
  }
  return false
}

type badRequestError struct {
  message string
}

func (e badRequestError) Error() string {
  return e.message
}

func errBadRequest(message string) error {
  return badRequestError{message: message}
}

func parsePagination(r *http.Request) (int, int, error) {
  page := 1
  pageSize := 20

  if raw := r.URL.Query().Get("page"); raw != "" {
    parsed, err := strconv.Atoi(raw)
    if err != nil || parsed <= 0 {
      return 0, 0, errBadRequest("invalid page")
    }
    page = parsed
  }

  if raw := r.URL.Query().Get("page_size"); raw != "" {
    parsed, err := strconv.Atoi(raw)
    if err != nil || parsed <= 0 {
      return 0, 0, errBadRequest("invalid page_size")
    }
    if parsed > 100 {
      parsed = 100
    }
    pageSize = parsed
  }

  return page, pageSize, nil
}
