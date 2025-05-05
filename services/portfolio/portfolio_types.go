package portfolio


type AddBalanceRequest struct {
	Amount float64 `json:"amount"`
}

type AssetQuantity struct {
	AssetName string    `json:"asset_name"`
	AssetID   string    `json:"asset_id"`
	Quantity  float64   `json:"quantity"`
	Symbol string `json:"symbol"`
	AssetType string `json:"asset_type"`
	PricePerUnit float64 `json:"price_per_unit"`
}