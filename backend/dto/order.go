package dto

import (
  "database/sql"
  "time"
)

// OrderDTO represents order fields read from database.
type OrderDTO struct {
  ID                 int64          `db:"id"`
  MerchantName       string         `db:"merchant_name"`
  TransactionNetwork string         `db:"transaction_network"`
  TransactionAsset   string         `db:"transaction_asset"`
  TXID               string         `db:"txid"`
  Amount             sql.NullFloat64 `db:"amount"`
  BeneficiaryName    string         `db:"beneficiary_name"`
  BankCountry        string         `db:"bank_country"`
  BankName           string         `db:"bank_name"`
  IBAN               string         `db:"iban"`
  SWIFT              string         `db:"swift"`
  ReferenceNote      sql.NullString  `db:"reference_note"`
  Status             string         `db:"status"`
  CreatedAt          time.Time      `db:"created_at"`
  UpdatedAt          time.Time      `db:"updated_at"`
}
