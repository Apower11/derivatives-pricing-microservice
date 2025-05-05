package messages

import(
	"testing"
	"github.com/Apower11/derivatives-pricing-microservice/services/auth"
	"github.com/Apower11/derivatives-pricing-microservice/utils"
	"github.com/Apower11/derivatives-pricing-microservice/types"
	"github.com/Apower11/derivatives-pricing-microservice/middleware"
	"github.com/Apower11/derivatives-pricing-microservice/services/chats"
	"net/http"
	"net/http/httptest"
	"bytes"
	"fmt"
	"errors"
	"gorm.io/gorm"
	"github.com/gorilla/mux"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/Apower11/derivatives-pricing-microservice/db"
)

func GetChatIDByName(name string) (string, error) {
    var chat types.MessageChat
    result := db.DB.Where("name = ?", name).Select("id").First(&chat)

    if result.Error != nil {
        if errors.Is(result.Error, gorm.ErrRecordNotFound) {
            return "", nil
        }
        return "", result.Error
    }
    return chat.ID, nil
}

func GetMessageIDByMessageContent(message_content string) (string, error) {
    var message Message
    result := db.DB.Where("message_content = ?", message_content).Select("id").First(&message)

    if result.Error != nil {
        if errors.Is(result.Error, gorm.ErrRecordNotFound) {
            return "", nil
        }
        return "", result.Error
    }
    return message.ID.String(), nil
}


var token string

func TestAddMessageHandler(t *testing.T){
	db.InitTestDB()
	_ = db.DB.Exec("DELETE FROM messages;")
	_ = db.DB.Exec("DELETE FROM message_chats;")
	_ = db.DB.Exec("DELETE FROM users;")
	r := mux.NewRouter()
	r.HandleFunc("/create-user", auth.CreateUserHandler)
	r.HandleFunc("/chat", chats.CreateChatHandler)
	protectedAddMessageHandler := middleware.AuthMiddleware(http.HandlerFunc(addMessageHandler))
	r.Handle("/message", protectedAddMessageHandler)

	w := httptest.NewRecorder()

	body := []byte(`{"username": "newuser", "password": "password123", "confirm_password": "password123", "is_admin": false}`)
	req := httptest.NewRequest("POST", "/create-user", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var responseMap map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &responseMap)
	token = responseMap["token"]

	body = []byte(`{"username": "newuser2", "password": "password123", "confirm_password": "password123", "is_admin": false}`)
	req = httptest.NewRequest("POST", "/create-user", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	firstUserExpectedUserId, err := utils.GetUserIDByUsername("newuser")
	if err != nil {
		t.Errorf("Error fetching user id: %v", err)
	}

	secondUserExpectedUserId, err := utils.GetUserIDByUsername("newuser2")
	if err != nil {
		t.Errorf("Error fetching user id: %v", err)
	}

	requestData := map[string]interface{}{
		"users_involved_in_chat":       []string{firstUserExpectedUserId, secondUserExpectedUserId},
		"name":        "newuser/newuser2",
		"type_of_chat": "direct_message_chat",
	}
	body, err = json.Marshal(requestData)
	req = httptest.NewRequest("POST", "/chat", bytes.NewBuffer(body))
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	expectedChatId, err := GetChatIDByName("newuser/newuser2")
	if err != nil {
		t.Errorf("Error chat user id: %v", err)
	}

	authHeader := fmt.Sprintf("Bearer %s", token)

	requestData = map[string]interface{}{
		"sender_id":       firstUserExpectedUserId,
		"chat_id":        expectedChatId,
		"message_content": "Example",
	}
	body, err = json.Marshal(requestData)
	req = httptest.NewRequest("POST", "/message", bytes.NewBuffer(body))
	req.Header.Set("Authorization", authHeader)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	expectedMessageId, err := GetMessageIDByMessageContent("Example")
	if err != nil {
		t.Errorf("Error getting message id: %v", err)
	}

	assert.Equal(t, http.StatusCreated, w.Code, "Status code should be Created")
	err = json.Unmarshal(w.Body.Bytes(), &responseMap)
	assert.Equal(t, expectedMessageId, responseMap["message_id"], "Should exist")
	assert.Equal(t, expectedChatId, responseMap["chat_id"], "Should exist")
	assert.Equal(t, firstUserExpectedUserId, responseMap["sender_id"], "Should exist")
	assert.Equal(t, "Example", responseMap["message_content"], "Chat ID should match")
}

func TestGetMessagesByChatHandler(t *testing.T) {
	r := mux.NewRouter()
	w := httptest.NewRecorder()
	protectedGetMessagesByChatHandler := middleware.AuthMiddleware(http.HandlerFunc(getMessagesByChatHandler))
	r.Handle("/chats/{chat_id}/messages", protectedGetMessagesByChatHandler)

	expectedChatId, err := GetChatIDByName("newuser/newuser2")
	if err != nil {
		t.Errorf("Error chat user id: %v", err)
	}

	authHeader := fmt.Sprintf("Bearer %s", token)
	requestUrl := fmt.Sprintf("/chats/%s/messages", expectedChatId)
	req := httptest.NewRequest("GET", requestUrl, nil)
	req.Header.Set("Authorization", authHeader)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	expectedMessageId, err := GetMessageIDByMessageContent("Example")
	if err != nil {
		t.Errorf("Error getting message id: %v", err)
	}

	var firstUserExpectedUserId string
	firstUserExpectedUserId, err = utils.GetUserIDByUsername("newuser")
	if err != nil {
		t.Errorf("Error fetching user id: %v", err)
	}

	body := w.Body.Bytes()

	var response chatWithMessagesResponse
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d. Response body: %s", http.StatusOK, w.Code, string(body))
		return
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		t.Errorf("Error unmarshalling JSON: %v. Response body: %s", err, string(body))
		return
	}

	assert.Equal(t, "newuser/newuser2", response.ChatName, "Test Case 1: Should return the correct chat name")
	assert.Equal(t, 1, len(response.Messages), "Test Case 1: Should return the correct number of messages")
	assert.Equal(t, firstUserExpectedUserId, response.Messages[0].SenderID, "Test Case 1: Should return the correct sender ID")
	assert.Equal(t, "newuser", response.Messages[0].SenderUsername, "Test Case 1: Should return the correct sender username")
	assert.Equal(t, "Example", response.Messages[0].MessageContent, "Test Case 1: Should return the correct message content")
	assert.Equal(t, expectedMessageId, response.Messages[0].ID.String(), "Test Case 1: Should return the correct message id")
}