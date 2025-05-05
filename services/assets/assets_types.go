package assets

type AddAssetRequest struct {
	AssetType string `json:"asset_type"`
	Symbol    string `json:"symbol"`
	Name      string `json:"name"`
}

type Asset struct {
    AssetID string `json:"asset_id"`
	AssetType string `json:"asset_type"`
	Symbol    string `json:"symbol"`
	Name      string `json:"name"`
}

type DBAsset struct {
    AssetID string `gorm:"column:asset_id"`
	Type     string  `gorm:"column:asset_type"`
    DerivativeID string `gorm:"column:derivative_id"`
	Name   string
	Symbol string
}