package types

import(
	"time"
	"github.com/lib/pq"
)

type Derivative struct {
	Name string `json:"name"`
	AssetName string `json:"asset_name"`
	SpotPrice   float64 `json:"spot_price"`
	StrikePrice float64 `json:"strike_price"`
	TimeToExpire float64 `json:"time_to_expire"`
	RiskFreeRate float64 `json:"risk_free_rate"`
	Volatility  float64 `json:"volatility"`
	Type        string  `json:"type"`
}

type PriceTimestamp struct {
	TimestampOfPrice  string `json:"timestamp"`
	Price   float64    `json:"price"`   
}

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

type User struct {
    ID        string `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Username string `gorm:"column:username;not null"`
	Password string `gorm:"column:password;not null"`
	CurrentBalance float64 `gorm:"column:current_balance;not null"`
    Category string `gorm:"column:category;not null"`
}

type Asset struct {
    AssetID string `json:"asset_id"`
	AssetType string `json:"asset_type"`
	Symbol    string `json:"symbol"`
	Name      string `json:"name"`
}


type MessageChat struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	ID                string   `gorm:"primaryKey;type:varchar(255)"`
    Name string     `gorm:"not null;type:varchar(255)"`
	UsersInvolvedInChat pq.StringArray `gorm:"type:text[]"`
	TypeOfChat        string   `gorm:"not null;type:varchar(255);check:type_of_chat IN ('direct_message_chat', 'group_chat')"`
}