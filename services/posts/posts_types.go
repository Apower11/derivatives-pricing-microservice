package posts

import(
	"time"
)

type Post struct {
	ID        string           `gorm:"primaryKey" json:"id"`
	UserID        string           `gorm:"not null" json:"user_id"`
	Text      string         `gorm:"not null" json:"text"`
	Image     []byte         `gorm:"type:bytea" json:"-"` 
	CreatedAt time.Time      `gorm:"not null" json:"created_at"`
	ImageBase64 string       `gorm:"-" json:"image"`      
}