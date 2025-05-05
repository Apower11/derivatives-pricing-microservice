package auth

type User struct {
    ID        string `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Username string `gorm:"column:username;not null"`
	Password string `gorm:"column:password;not null"`
	CurrentBalance float64 `gorm:"column:current_balance;not null"`
    Category string `gorm:"column:category;not null"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
	Username string `json:"username"`
	ID string `json:"id"`
	Category string `json:"category"`
}

type UserRequest struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
    IsAdmin         bool `json:"is_admin"`
}