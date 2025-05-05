package users

type SearchUser struct {
    ID        string `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Username string `gorm:"column:username;not null"`
}