package derivatives

import(
	"time"
)

type DBDerivative struct {
	DerivativeID   string       `gorm:"column:id;primaryKey"`
	DerivativeName string    `gorm:"column:derivative_name;not null"`
	DerivativeType string    `gorm:"column:derivative_type;not null"`
	AssetName      string    `gorm:"column:asset_name;not null"`
	StrikePrice    float64   `gorm:"column:strike_price;not null"`
	ExpirationDate time.Time `gorm:"column:expiration_date;not null"`
	RiskFreeRate   float64   `gorm:"column:risk_free_rate;not null"`
    TimeToExpiry   float64   `gorm:"-"`
    CurrentPrice   float64   `gorm:"-"`
	SpotPrice float64 `gorm:"-"`
	Volatility     float64   `gorm:"-"`
}

type PriceTimestamp struct {
	TimestampOfPrice  string `json:"timestamp"`
	Price   float64    `json:"price"`   
}

type AddDerivativeRequestData struct {
	Name  string `json:"name"`
	AssetName   string    `json:"assetName"`
	DerivativeType string `json:"derivativeType"`
	StrikePrice float64 `json:"strikePrice"`
	ExpirationDate string `json:"expirationDate"`
	RiskFreeRate float64 `json:"riskFreeRate"`
	Symbol string `json:"symbol"`
}

type priceResponse struct {
	Prices []float64 `json:"prices"`
}

type priceObject struct {
	Close float64 `json:"close"`
}

type AssetVolume struct {
	Volume  float64     `gorm:"column:volume;not null`
}