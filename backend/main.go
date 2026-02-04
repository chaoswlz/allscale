package main

import (
  "context"
  "database/sql"
  "fmt"
  "log"
  "net/http"
  "os"
  "strconv"
  "strings"
  "time"

  _ "github.com/go-sql-driver/mysql"
  "sarah-project-backend/handler"
)

func main() {
  db, err := openDB()
  if err != nil {
    log.Fatal(err)
  }
  defer db.Close()
  if err := ensureTables(db); err != nil {
    log.Fatal(err)
  }

  jwtConfig, err := loadJWTConfig()
  if err != nil {
    log.Fatal(err)
  }

  customerConfig, err := loadCustomerAuthConfig()
  if err != nil {
    log.Fatal(err)
  }

  mux := http.NewServeMux()
  mux.HandleFunc("/admin/login", handler.AdminLogin(db, jwtConfig))
  mux.HandleFunc("/admin/stats", handler.AdminStats(db, jwtConfig))
  mux.HandleFunc("/admin/ready-processing", handler.AdminReadyProcessing(db, jwtConfig))
  mux.HandleFunc("/admin/recent-orders", handler.AdminRecentOrders(db, jwtConfig))
  mux.HandleFunc("/admin/order", handler.AdminOrderDetail(db, jwtConfig))
  mux.HandleFunc("/admin/order/status", handler.AdminUpdateOrderStatus(db, jwtConfig))
  mux.HandleFunc("/customer/createOrder", handler.CreateOrder(db, customerConfig))
  mux.HandleFunc("/customer/orders", handler.ListCustomerOrders(db, customerConfig))
  mux.HandleFunc("/customer/order", handler.GetCustomerOrder(db, customerConfig))
  mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    _, _ = w.Write([]byte(`{"status":"ok"}`))
  })

  server := &http.Server{
    Addr:         ":8080",
    Handler:      mux,
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  60 * time.Second,
  }

  log.Println("API listening on :8080")
  if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
    log.Fatal(err)
  }
}

func openDB() (*sql.DB, error) {
  host := os.Getenv("MYSQL_HOST")
  port := os.Getenv("MYSQL_PORT")
  user := os.Getenv("MYSQL_USER")
  password := os.Getenv("MYSQL_PASSWORD")
  dbName := os.Getenv("MYSQL_DB")
  params := os.Getenv("MYSQL_PARAMS")
  if params == "" {
    params = "charset=utf8mb4&parseTime=True&loc=Local"
  }

  missing := make([]string, 0, 5)
  if host == "" {
    missing = append(missing, "MYSQL_HOST")
  }
  if port == "" {
    missing = append(missing, "MYSQL_PORT")
  }
  if user == "" {
    missing = append(missing, "MYSQL_USER")
  }
  if password == "" {
    missing = append(missing, "MYSQL_PASSWORD")
  }
  if dbName == "" {
    missing = append(missing, "MYSQL_DB")
  }
  if len(missing) > 0 {
    return nil, fmt.Errorf("missing env vars: %s (see .env.example)", strings.Join(missing, ", "))
  }

  dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s", user, password, host, port, dbName, params)

  db, err := sql.Open("mysql", dsn)
  if err != nil {
    return nil, err
  }

  ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
  defer cancel()
  if err := db.PingContext(ctx); err != nil {
    return nil, err
  }

  return db, nil
}

func loadJWTConfig() (handler.AuthConfig, error) {
  secret := os.Getenv("JWT_SECRET")
  if secret == "" {
    return handler.AuthConfig{}, fmt.Errorf("JWT_SECRET is required (see .env.example)")
  }

  issuer := os.Getenv("JWT_ISSUER")
  if issuer == "" {
    issuer = "sarah-project"
  }

  ttlMinutes := 60
  if raw := os.Getenv("JWT_TTL_MINUTES"); raw != "" {
    parsed, err := strconv.Atoi(raw)
    if err != nil || parsed <= 0 {
      return handler.AuthConfig{}, fmt.Errorf("JWT_TTL_MINUTES must be a positive integer")
    }
    ttlMinutes = parsed
  }

  return handler.AuthConfig{
    JWTSecret: secret,
    JWTIssuer: issuer,
    JWTTTL:    time.Duration(ttlMinutes) * time.Minute,
  }, nil
}

func loadCustomerAuthConfig() (handler.CustomerAuthConfig, error) {
  apiKey := os.Getenv("CUSTOMER_API_KEY")
  if apiKey == "" {
    return handler.CustomerAuthConfig{}, fmt.Errorf("CUSTOMER_API_KEY is required (see .env.example)")
  }

  merchantName := os.Getenv("CUSTOMER_MERCHANT_NAME")
  if merchantName == "" {
    return handler.CustomerAuthConfig{}, fmt.Errorf("CUSTOMER_MERCHANT_NAME is required (see .env.example)")
  }

  return handler.CustomerAuthConfig{
    APIKey:       apiKey,
    MerchantName: merchantName,
  }, nil
}

func ensureTables(db *sql.DB) error {
  ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
  defer cancel()

  statements := []string{
    `
      CREATE TABLE IF NOT EXISTS admin_users (
        id BIGINT AUTO_INCREMENT PRIMARY KEY,
        username VARCHAR(64) NOT NULL UNIQUE,
        email VARCHAR(128) NOT NULL,
        password_hash VARCHAR(255) NOT NULL,
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
      ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
    `,
    `
      CREATE TABLE IF NOT EXISTS customer_users (
        id BIGINT AUTO_INCREMENT PRIMARY KEY,
        name VARCHAR(128) NOT NULL,
        email VARCHAR(128) NOT NULL UNIQUE,
        password_hash VARCHAR(255) NOT NULL,
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
      ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
    `,
    `
      CREATE TABLE IF NOT EXISTS orders (
        id BIGINT AUTO_INCREMENT PRIMARY KEY,
        merchant_name VARCHAR(128) NOT NULL,
        transaction_network ENUM('TRON', 'BSC', 'Ethereum') NOT NULL,
        transaction_asset ENUM('USDT', 'USDC') NOT NULL,
        txid VARCHAR(128) NOT NULL,
        amount DECIMAL(18, 8) NULL,
        beneficiary_name VARCHAR(128) NOT NULL,
        bank_country ENUM('Canada', 'United States') NOT NULL,
        bank_name VARCHAR(128) NOT NULL,
        iban VARCHAR(64) NOT NULL,
        swift VARCHAR(64) NOT NULL,
        reference_note TEXT NULL,
        status ENUM('Paid', 'Processing', 'Summitted', 'Failed', 'Funds Received') NOT NULL DEFAULT 'Processing',
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
      ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
    `,
  }

  for _, stmt := range statements {
    if _, err := db.ExecContext(ctx, stmt); err != nil {
      return err
    }
  }

  return nil
}
