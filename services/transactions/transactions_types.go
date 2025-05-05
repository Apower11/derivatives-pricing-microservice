package transactions

import(
	"time"
)

type Transaction struct {
	TransactionID   string       `json:"transaction_id"`
	UserID          string       `json:"user_id"`
	AssetID         string       `json:"asset_id"`
	TransactionType string    `json:"transaction_type"`
	TransactionDate time.Time `json:"transaction_date"`
	Quantity        float64   `json:"quantity"`
	PricePerUnit    float64   `json:"price_per_unit"`
}

type CreateTransactionRequest struct {
	AssetID      string     `json:"asset_id"`
	Quantity     float64 `json:"quantity"`
	PricePerUnit float64 `json:"price_per_unit"`
}

type CreateSellTransactionRequest struct {
	AssetID      string  `json:"asset_id"`
	Quantity     float64 `json:"quantity"`
	PricePerUnit float64 `json:"price_per_unit"`
}