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

type adminStats struct {
  FundsReceived int64 `json:"funds_received"`
  Processing    int64 `json:"processing"`
  ActionRequired int64 `json:"action_required"`
  Awaiting      int64 `json:"awaiting"`
  CompletedToday int64 `json:"completed_today"`
}

type adminOrderRow struct {
  OrderID      int64     `json:"order_id"`
  MerchantName string    `json:"merchant_name"`
  Asset        string    `json:"asset"`
  Network      string    `json:"network"`
  Amount       *float64  `json:"amount"`
  TimeReceived time.Time `json:"time_received"`
}

type adminRecentRow struct {
  OrderID      int64     `json:"order_id"`
  Status       string    `json:"status"`
  MerchantName string    `json:"merchant_name"`
  Network      string    `json:"network"`
  Amount       *float64  `json:"amount"`
  Asset        string    `json:"asset"`
  LastUpdate   time.Time `json:"last_update"`
}

type adminOrderDetail struct {
  OrderID           int64     `json:"order_id"`
  MerchantName      string    `json:"merchant_name"`
  TransactionNetwork string    `json:"transaction_network"`
  TransactionAsset   string    `json:"transaction_asset"`
  TXID              string    `json:"txid"`
  Amount            *float64  `json:"amount"`
  BeneficiaryName   string    `json:"beneficiary_name"`
  BankCountry       string    `json:"bank_country"`
  BankName          string    `json:"bank_name"`
  IBAN              string    `json:"iban"`
  SWIFT             string    `json:"swift"`
  ReferenceNote     *string   `json:"reference_note"`
  Status            string    `json:"status"`
  CreatedAt         time.Time `json:"created_at"`
}

type adminOrderDetailResponse struct {
  Order adminOrderDetail `json:"order"`
}

type updateOrderStatusRequest struct {
  ID     int64  `json:"id"`
  Status string `json:"status"`
}

type adminListResponse[T any] struct {
  Total    int64 `json:"total"`
  Page     int   `json:"page"`
  PageSize int   `json:"page_size"`
  Items    []T   `json:"items"`
}

// AdminStats returns summary stats for admin UI.
func AdminStats(db *sql.DB, cfg AuthConfig) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
      writeError(w, http.StatusMethodNotAllowed, "method not allowed")
      return
    }

    claims, err := authenticateRequest(r, cfg.JWTSecret)
    if err != nil || claims.Role != "admin" {
      writeError(w, http.StatusUnauthorized, "unauthorized")
      return
    }

    stats, err := loadAdminStats(r.Context(), db)
    if err != nil {
      log.Printf("admin stats error: %v", err)
      writeError(w, http.StatusInternalServerError, "server error")
      return
    }

    writeJSON(w, http.StatusOK, stats)
  }
}

// AdminReadyProcessing returns processing orders with pagination.
func AdminReadyProcessing(db *sql.DB, cfg AuthConfig) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
      writeError(w, http.StatusMethodNotAllowed, "method not allowed")
      return
    }

    claims, err := authenticateRequest(r, cfg.JWTSecret)
    if err != nil || claims.Role != "admin" {
      writeError(w, http.StatusUnauthorized, "unauthorized")
      return
    }

    page, pageSize, err := parsePagination(r)
    if err != nil {
      writeError(w, http.StatusBadRequest, err.Error())
      return
    }

    total, rows, err := loadReadyProcessing(r.Context(), db, page, pageSize)
    if err != nil {
      log.Printf("admin ready list error: %v", err)
      writeError(w, http.StatusInternalServerError, "server error")
      return
    }

    writeJSON(w, http.StatusOK, adminListResponse[adminOrderRow]{
      Total:    total,
      Page:     page,
      PageSize: pageSize,
      Items:    rows,
    })
  }
}

// AdminRecentOrders returns recent orders with pagination.
func AdminRecentOrders(db *sql.DB, cfg AuthConfig) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
      writeError(w, http.StatusMethodNotAllowed, "method not allowed")
      return
    }

    claims, err := authenticateRequest(r, cfg.JWTSecret)
    if err != nil || claims.Role != "admin" {
      writeError(w, http.StatusUnauthorized, "unauthorized")
      return
    }

    page, pageSize, err := parsePagination(r)
    if err != nil {
      writeError(w, http.StatusBadRequest, err.Error())
      return
    }

    total, rows, err := loadRecentOrders(r.Context(), db, page, pageSize)
    if err != nil {
      log.Printf("admin recent list error: %v", err)
      writeError(w, http.StatusInternalServerError, "server error")
      return
    }

    writeJSON(w, http.StatusOK, adminListResponse[adminRecentRow]{
      Total:    total,
      Page:     page,
      PageSize: pageSize,
      Items:    rows,
    })
  }
}

// AdminOrderDetail returns a single order by ID for admin view.
func AdminOrderDetail(db *sql.DB, cfg AuthConfig) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
      writeError(w, http.StatusMethodNotAllowed, "method not allowed")
      return
    }

    claims, err := authenticateRequest(r, cfg.JWTSecret)
    if err != nil || claims.Role != "admin" {
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

    order, err := loadAdminOrderDetail(r.Context(), db, orderID)
    if err != nil {
      if err == sql.ErrNoRows {
        writeError(w, http.StatusNotFound, "order not found")
        return
      }
      log.Printf("admin order detail error: %v", err)
      writeError(w, http.StatusInternalServerError, "server error")
      return
    }

    writeJSON(w, http.StatusOK, adminOrderDetailResponse{Order: order})
  }
}

// AdminUpdateOrderStatus updates an order status.
func AdminUpdateOrderStatus(db *sql.DB, cfg AuthConfig) http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
      writeError(w, http.StatusMethodNotAllowed, "method not allowed")
      return
    }

    claims, err := authenticateRequest(r, cfg.JWTSecret)
    if err != nil || claims.Role != "admin" {
      writeError(w, http.StatusUnauthorized, "unauthorized")
      return
    }

    var req updateOrderStatusRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
      writeError(w, http.StatusBadRequest, "invalid json body")
      return
    }
    if req.ID <= 0 {
      writeError(w, http.StatusBadRequest, "invalid id")
      return
    }
    if !isAllowedStatus(req.Status) {
      writeError(w, http.StatusBadRequest, "invalid status")
      return
    }

    order, err := updateOrderStatus(r.Context(), db, req.ID, req.Status)
    if err != nil {
      if err == sql.ErrNoRows {
        writeError(w, http.StatusNotFound, "order not found")
        return
      }
      log.Printf("admin update status error: %v", err)
      writeError(w, http.StatusInternalServerError, "server error")
      return
    }

    writeJSON(w, http.StatusOK, adminOrderDetailResponse{Order: order})
  }
}

func loadAdminStats(ctx context.Context, db *sql.DB) (adminStats, error) {
  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()

  today := time.Now().Format("2006-01-02")
  stats := adminStats{}

  queries := []struct {
    sql    string
    target *int64
  }{
    {`SELECT COUNT(*) FROM orders WHERE status = 'Funds Received'`, &stats.FundsReceived},
    {`SELECT COUNT(*) FROM orders WHERE status = 'Processing'`, &stats.Processing},
    {`SELECT COUNT(*) FROM orders WHERE status = 'Failed'`, &stats.ActionRequired},
    {`SELECT COUNT(*) FROM orders WHERE status = 'Summitted'`, &stats.Awaiting},
    {`SELECT COUNT(*) FROM orders WHERE status = 'Paid' AND DATE(created_at) = ?`, &stats.CompletedToday},
  }

  for _, q := range queries {
    if q.sql == "SELECT COUNT(*) FROM orders WHERE status = 'Paid' AND DATE(created_at) = ?" {
      if err := db.QueryRowContext(ctx, q.sql, today).Scan(q.target); err != nil {
        return adminStats{}, err
      }
      continue
    }
    if err := db.QueryRowContext(ctx, q.sql).Scan(q.target); err != nil {
      return adminStats{}, err
    }
  }

  return stats, nil
}

func loadReadyProcessing(ctx context.Context, db *sql.DB, page int, pageSize int) (int64, []adminOrderRow, error) {
  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()

  var total int64
  if err := db.QueryRowContext(ctx, `
    SELECT COUNT(*) FROM orders WHERE status = 'Processing'
  `).Scan(&total); err != nil {
    return 0, nil, err
  }

  offset := (page - 1) * pageSize
  rows, err := db.QueryContext(ctx, `
    SELECT id, merchant_name, transaction_asset, transaction_network, amount, created_at
    FROM orders
    WHERE status = 'Processing'
    ORDER BY created_at DESC
    LIMIT ? OFFSET ?
  `, pageSize, offset)
  if err != nil {
    return 0, nil, err
  }
  defer rows.Close()

  results := make([]adminOrderRow, 0)
  for rows.Next() {
    var (
      amount sql.NullFloat64
      row    adminOrderRow
    )
    if err := rows.Scan(
      &row.OrderID,
      &row.MerchantName,
      &row.Asset,
      &row.Network,
      &amount,
      &row.TimeReceived,
    ); err != nil {
      return 0, nil, err
    }
    if amount.Valid {
      row.Amount = &amount.Float64
    }
    results = append(results, row)
  }
  if err := rows.Err(); err != nil {
    return 0, nil, err
  }

  return total, results, nil
}

func loadRecentOrders(ctx context.Context, db *sql.DB, page int, pageSize int) (int64, []adminRecentRow, error) {
  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()

  var total int64
  if err := db.QueryRowContext(ctx, `
    SELECT COUNT(*) FROM orders
  `).Scan(&total); err != nil {
    return 0, nil, err
  }

  offset := (page - 1) * pageSize
  rows, err := db.QueryContext(ctx, `
    SELECT id, status, merchant_name, transaction_network, amount, transaction_asset, updated_at
    FROM orders
    ORDER BY updated_at DESC
    LIMIT ? OFFSET ?
  `, pageSize, offset)
  if err != nil {
    return 0, nil, err
  }
  defer rows.Close()

  results := make([]adminRecentRow, 0)
  for rows.Next() {
    var (
      amount sql.NullFloat64
      row    adminRecentRow
    )
    if err := rows.Scan(
      &row.OrderID,
      &row.Status,
      &row.MerchantName,
      &row.Network,
      &amount,
      &row.Asset,
      &row.LastUpdate,
    ); err != nil {
      return 0, nil, err
    }
    if amount.Valid {
      row.Amount = &amount.Float64
    }
    results = append(results, row)
  }
  if err := rows.Err(); err != nil {
    return 0, nil, err
  }

  return total, results, nil
}

func loadAdminOrderDetail(ctx context.Context, db *sql.DB, orderID int64) (adminOrderDetail, error) {
  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()

  var (
    amount sql.NullFloat64
    note   sql.NullString
    order  adminOrderDetail
  )

  row := db.QueryRowContext(ctx, `
    SELECT id, merchant_name, transaction_network, transaction_asset, txid, amount,
           beneficiary_name, bank_country, bank_name, iban, swift, reference_note, status, created_at
    FROM orders
    WHERE id = ?
    LIMIT 1
  `, orderID)

  if err := row.Scan(
    &order.OrderID,
    &order.MerchantName,
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
    return adminOrderDetail{}, err
  }

  if amount.Valid {
    order.Amount = &amount.Float64
  }
  if note.Valid && strings.TrimSpace(note.String) != "" {
    order.ReferenceNote = &note.String
  }

  return order, nil
}

func updateOrderStatus(ctx context.Context, db *sql.DB, orderID int64, status string) (adminOrderDetail, error) {
  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()

  res, err := db.ExecContext(ctx, `
    UPDATE orders
    SET status = ?
    WHERE id = ?
  `, status, orderID)
  if err != nil {
    return adminOrderDetail{}, err
  }
  if rows, err := res.RowsAffected(); err != nil || rows == 0 {
    if err != nil {
      return adminOrderDetail{}, err
    }
    return adminOrderDetail{}, sql.ErrNoRows
  }

  return loadAdminOrderDetail(ctx, db, orderID)
}

func isAllowedStatus(status string) bool {
  switch status {
  case "Paid", "Processing", "Summitted", "Failed", "Funds Received":
    return true
  default:
    return false
  }
}
