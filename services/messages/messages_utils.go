package messages

import(
	"sync"
	"encoding/json"
	"fmt"
	"log"
	"gorm.io/gorm"
	"github.com/gorilla/websocket"
	"github.com/Apower11/derivatives-pricing-microservice/types"
	"github.com/Apower11/derivatives-pricing-microservice/db"
	"time"
	"github.com/google/uuid"
	"net/http"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var connections = make(map[string][]*websocket.Conn)
var connectionsMutex = &sync.Mutex{}

func GetUserById(id string) (types.User, error) {
	var user types.User
	result := db.DB.Where("id = ?", id).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return types.User{}, fmt.Errorf("user not found: %w", result.Error)
		}
		return types.User{}, fmt.Errorf("failed to query user: %w", result.Error)
	}
	return user, nil
}

func removeConnection(chatID string, conn *websocket.Conn) {
    connectionsMutex.Lock() 
	defer connectionsMutex.Unlock()
	chatConnections := connections[chatID]
	if chatConnections == nil || len(chatConnections) == 0 {
		return
	}

	for i, c := range chatConnections {
		if c == conn {
			connections[chatID] = append(chatConnections[:i], chatConnections[i+1:]...)
			break
		}
	}
	if len(connections[chatID]) == 0{
		delete(connections, chatID)
	}

}

func notifyChat(chatID string, message Message) {
	connectionsMutex.Lock() 
	chatConnections := connections[chatID]
	connectionsMutex.Unlock() 
	if chatConnections == nil || len(chatConnections) == 0 {
		return 
	}

	type websocketMessage struct {
		ID            uuid.UUID `json:"id"`
		ChatID        string    `json:"chat_id"`
		SenderID      string    `json:"sender_id"`
		SenderUsername string    `json:"sender_username"`
		MessageContent  string    `json:"message_content"`
		Timestamp     time.Time `json:"timestamp"`
	}
	user, err := GetUserById(message.IDOfSender)
	if err != nil{
		log.Printf("Error getting user: %v", err)
		return
	}

	wsMessage := websocketMessage{
		ID:            message.ID,
		ChatID:        message.ChatID,
		SenderID:      message.IDOfSender,
		SenderUsername: user.Username,
		MessageContent:  message.MessageContent,
		Timestamp:     message.Timestamp,
	}

	messageJSON, err := json.Marshal(wsMessage)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	for _, conn := range chatConnections {
		err := conn.WriteMessage(websocket.TextMessage, messageJSON)
		if err != nil {
			log.Printf("Error sending message to client: %v", err)
			removeConnection(chatID, conn)
			conn.Close()
		}
	}
}