package messages

import (
	"github.com/Apower11/derivatives-pricing-microservice/types"
	"time"
	"github.com/google/uuid"
)

type Message struct {
	ID          uuid.UUID `gorm:"primaryKey;type:uuid"`
	ChatID      string    `gorm:"not null;type:varchar(255)"`
	IDOfSender    string    `gorm:"not null;type:varchar(255)"`
	MessageContent string    `gorm:"not null;type:text"`
	Timestamp   time.Time `gorm:"not null;default:now()"` 
	Chat        types.MessageChat `gorm:"foreignKey:ChatID"`
	Sender      types.User      `gorm:"foreignKey:IDOfSender"`   
}

type addMessageRequest struct {
	SenderID    string `json:"sender_id"`
	ChatID      string `json:"chat_id"`
	MessageContent string `json:"message_content"`
}

type messageResponse struct {
	ID             uuid.UUID `json:"id"`
	ChatID         string    `json:"chat_id"`
	SenderID       string    `json:"sender_id"`
	SenderUsername string    `json:"sender_username"`
	MessageContent string    `json:"message_content"`
	Timestamp      time.Time `json:"timestamp"`
}

type chatWithMessagesResponse struct {
	ChatName string            `json:"chat_name"`
	Messages []messageResponse `json:"messages"`
}