package messages

import(
	"encoding/json"
	"fmt"
	"net/http"
	"log"
	"gorm.io/gorm"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"errors"
	"github.com/Apower11/derivatives-pricing-microservice/db"
	"github.com/Apower11/derivatives-pricing-microservice/utils"
	"github.com/Apower11/derivatives-pricing-microservice/types"
)

func addMessageHandler(w http.ResponseWriter, r *http.Request) {
	utils.EnableCORS(&w)
	if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
	}

	var err error
        
	var req addMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Error decoding request body: %v", err), http.StatusBadRequest)
		return
	}

	if req.SenderID == "" || req.ChatID == "" || req.MessageContent == "" {
		http.Error(w, "sender_id, chat_id, and message_content are required", http.StatusBadRequest)
		return
	}

	newMessage := Message{
		ID:          uuid.New(),
		ChatID:      req.ChatID,
		IDOfSender:    req.SenderID,
		MessageContent: req.MessageContent,
	}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		result := tx.Create(&newMessage)
		if result.Error != nil {
			return fmt.Errorf("failed to insert new message: %w", result.Error)
		}
		if result.RowsAffected != 1 {
			return fmt.Errorf("expected 1 row to be affected, got %d", result.RowsAffected)
		}

		return nil
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to add message: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := map[string]string{
		"message_id":   newMessage.ID.String(),
		"chat_id":      newMessage.ChatID,
		"sender_id":    newMessage.IDOfSender,
		"message_content": newMessage.MessageContent,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
    notifyChat(newMessage.ChatID, newMessage)
}

func getMessagesByChatHandler(w http.ResponseWriter, r *http.Request) {
	utils.EnableCORS(&w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	vars := mux.Vars(r)
	chatID := vars["chat_id"]

	if chatID == "" {
		http.Error(w, "chat_id is required", http.StatusBadRequest)
		return
	}

	var chat types.MessageChat 
	chatResult := db.DB.First(&chat, "id = ?", chatID)
	if chatResult.Error != nil {
		if errors.Is(chatResult.Error, gorm.ErrRecordNotFound) {
			http.Error(w, fmt.Sprintf("Chat not found with ID: %s", chatID), http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to fetch chat: %v", chatResult.Error), http.StatusInternalServerError)
		}
		return
	}

	var messages []Message
	result := db.DB.
		Preload("Sender"). 
		Where("chat_id = ?", chatID).
		Order("timestamp ASC"). 
		Find(&messages)

	if result.Error != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch messages: %v", result.Error), http.StatusInternalServerError)
		return
	}

	var responseMessages []messageResponse
	for _, message := range messages {
		responseMessages = append(responseMessages, messageResponse{
			ID:             message.ID,
			ChatID:         message.ChatID,
			SenderID:       message.IDOfSender,
			SenderUsername: message.Sender.Username, // Access the username from the loaded Sender
			MessageContent: message.MessageContent,
			Timestamp:      message.Timestamp,
		})
	}

	finalResponse := chatWithMessagesResponse{
		ChatName: chat.Name,
		Messages: responseMessages,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(finalResponse); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
		return
	}
	log.Printf("Fetched %d messages for chat ID: %s, Chat Name: %s", len(messages), chatID, chat.Name)
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	chatID := vars["chat_id"]

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	connectionsMutex.Lock()       
	connections[chatID] = append(connections[chatID], conn)
	connectionsMutex.Unlock() 

	defer func() {
		removeConnection(chatID, conn)
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
	}
}