package chats

import (
	"fmt"
	"net/http"
	"github.com/Apower11/derivatives-pricing-microservice/utils"
	"github.com/Apower11/derivatives-pricing-microservice/types"
	"github.com/Apower11/derivatives-pricing-microservice/db"
	"encoding/json"
	"log"
	"gorm.io/gorm"
	"github.com/lib/pq"
	"github.com/google/uuid"
)

func CreateChatHandler(w http.ResponseWriter, r *http.Request) {
	utils.EnableCORS(&w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	var err error

	var req createChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Error decoding request body: %v", err), http.StatusBadRequest)
		return
	}

	if req.TypeOfChat != "direct_message_chat" && req.TypeOfChat != "group_chat" {
		http.Error(w, "Invalid type_of_chat. Must be 'direct_message_chat' or 'group_chat'", http.StatusBadRequest)
		return
	}
	if len(req.UsersInvolvedInChat) < 2 {
		http.Error(w, "At least two users are required to create a chat", http.StatusBadRequest)
		return
	}

	newChat := types.MessageChat{
		ID: uuid.New().String(), 
        Name: req.Name,
		TypeOfChat:        req.TypeOfChat,
		UsersInvolvedInChat: req.UsersInvolvedInChat,
	}

	err = db.DB.Transaction(func(tx *gorm.DB) error {
		result := tx.Exec(`
			INSERT INTO message_chats (id, type_of_chat, users_involved_in_chat, name)
			VALUES (?, ?, ?, ?)
		`, newChat.ID, newChat.TypeOfChat, pq.Array(req.UsersInvolvedInChat), newChat.Name)

		if result.Error != nil {
			return fmt.Errorf("failed to insert into message_chats: %w", result.Error) 
		}

		if result.RowsAffected != 1 {
			return fmt.Errorf("expected 1 row to be affected, got %d", result.RowsAffected)
		}

		return nil 
	})

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create chat: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := map[string]string{"chat_id": newChat.ID}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
	log.Printf("Chat created with ID: %s", newChat.ID)
}

func getUserChatsHandler(w http.ResponseWriter, r *http.Request) {
	utils.EnableCORS(&w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "User ID not found in request context", http.StatusInternalServerError)
		return
	}

	var chats []types.MessageChat
	result := db.DB.Raw(`
        SELECT id, type_of_chat, users_involved_in_chat, name
        FROM message_chats
        WHERE ? = ANY(users_involved_in_chat)
    `, userID).Scan(&chats)
	if result.Error != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch chats: %v", result.Error), http.StatusInternalServerError)
		return
	}

	var chatIDs []string
	for _, chat := range chats {
		chatIDs = append(chatIDs, chat.ID)
	}

	var mostRecentMessages []struct {
		ChatID        string `gorm:"column:chat_id"`
		MessageContent string `gorm:"column:message_content"`
		SenderUsername string `gorm:"column:sender_username"` // Added field
	}

	if len(chatIDs) > 0 {
		result = db.DB.Raw(`
            SELECT
                m.chat_id,
                m.message_content,
                u.username AS sender_username 
            FROM (
                SELECT
                    chat_id,
                    message_content,
                    id_of_sender,
                    timestamp,
                    ROW_NUMBER() OVER(PARTITION BY chat_id ORDER BY timestamp DESC) as rn
                FROM messages
                WHERE chat_id IN (?)
            ) m
            JOIN users u ON m.id_of_sender = u.id  -- Join with the users table
            WHERE m.rn = 1
        `, chatIDs).Scan(&mostRecentMessages)

		if result.Error != nil {
			log.Printf("Error fetching most recent messages and senders: %v", result.Error)
		}
	}

	type chatResponse struct {
		ID              string   `json:"id"`
		Name            string   `json:"name"`
		MostRecentMessage string `json:"most_recent_message"`
		SenderUsername    string `json:"sender_username"`    // Added field
		TypeOfChat      string   `json:"type_of_chat"`
		UsersInvolved   []string `json:"users_involved_in_chat"`
	}

	var response []chatResponse
	mostRecentMessageMap := make(map[string]struct {
		Content  string
		Username string
	})
	for _, msg := range mostRecentMessages {
		mostRecentMessageMap[msg.ChatID] = struct {
			Content  string
			Username string
		}{Content: msg.MessageContent, Username: msg.SenderUsername}
	}

	for _, chat := range chats {
		mostRecent := mostRecentMessageMap[chat.ID] 
		response = append(response, chatResponse{
			ID:              chat.ID,
			Name:            chat.Name,
			MostRecentMessage: mostRecent.Content,
			SenderUsername:    mostRecent.Username,
			TypeOfChat:      chat.TypeOfChat,
			UsersInvolved:   chat.UsersInvolvedInChat,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
		return
	}
	log.Printf("Successfully fetched %d chats for user ID: %s", len(chats), userID)
}

func checkChatExistsHandler(w http.ResponseWriter, r *http.Request) {
	utils.EnableCORS(&w)
	if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
	}

    var err error
	userIDFromToken, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "User ID not found in request context", http.StatusInternalServerError)
		return
	}

	var req CheckChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Error decoding request body: %v", err), http.StatusBadRequest)
		return
	}
	otherUserID := req.UserID

	if otherUserID == "" {
		http.Error(w, "User ID is required in the request body", http.StatusBadRequest)
		return
	}

	var chat types.MessageChat
	err = db.DB.
		Where("type_of_chat = ?", "direct_message_chat").
		Where("? = ANY(users_involved_in_chat)", userIDFromToken).
		Where("? = ANY(users_involved_in_chat)", otherUserID).
		First(&chat).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := map[string]bool{"exists": false}
			if err := json.NewEncoder(w).Encode(response); err != nil {
				log.Printf("Error encoding response: %v", err)
			}
			return
		}
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]interface{}{"exists": true, "chat_id": chat.ID}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
	log.Printf("Chat exists with ID: %s", chat.ID)
}

