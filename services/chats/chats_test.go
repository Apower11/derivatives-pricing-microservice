package chats

import(
	"testing"
	"github.com/Apower11/derivatives-pricing-microservice/services/auth"
	"github.com/Apower11/derivatives-pricing-microservice/utils"
	"github.com/Apower11/derivatives-pricing-microservice/middleware"
	"net/http"
	"net/http/httptest"
	"bytes"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"github.com/gorilla/mux"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/Apower11/derivatives-pricing-microservice/types"
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

var token string

func TestCreateChatHandler(t *testing.T){
	db.InitTestDB()
	_ = db.DB.Exec("DELETE FROM messages;")
	_ = db.DB.Exec("DELETE FROM message_chats;")
	_ = db.DB.Exec("DELETE FROM users;")
	body := []byte(`{"username": "newuser", "password": "password123", "confirm_password": "password123", "is_admin": false}`)
	req := httptest.NewRequest("POST", "/create-user", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/create-user", auth.CreateUserHandler)
	r.ServeHTTP(w, req)
	firstUserExpectedUserId, err := utils.GetUserIDByUsername("newuser")
	if err != nil {
		t.Errorf("Error fetching user id: %v", err)
	}

	var responseMap map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &responseMap)
	token = responseMap["token"]

	w = httptest.NewRecorder()
	body = []byte(`{"username": "newuser2", "password": "password123", "confirm_password": "password123", "is_admin": true}`)
	req = httptest.NewRequest("POST", "/create-user", bytes.NewBuffer(body))
	r.ServeHTTP(w, req)
	secondUserExpectedUserId, err := utils.GetUserIDByUsername("newuser2")
	if err != nil {
		t.Errorf("Error fetching user id: %v", err)
	}

	r.HandleFunc("/chat", CreateChatHandler)
	body = []byte(`{"Name": "newuser2", "password": "password123", "confirm_password": "password123", "is_admin": true}`)
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

	assert.Equal(t, http.StatusCreated, w.Code, "Status code should be Created")
	err = json.Unmarshal(w.Body.Bytes(), &responseMap)
	assert.NoError(t, err, "Should not error unmarshalling JSON")
	assert.Equal(t, expectedChatId, responseMap["chat_id"], "Chat ids should match")
}

func TestGetUserChatsHandler(t *testing.T){
	authHeader := fmt.Sprintf("Bearer %s", token)
	req := httptest.NewRequest("GET", "/chats", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)

	w := httptest.NewRecorder()
	r := mux.NewRouter()
	protectedGetUserChatsHandler := middleware.AuthMiddleware(http.HandlerFunc(getUserChatsHandler))
	r.Handle("/chats", protectedGetUserChatsHandler)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code, "Status code should be OK")
	var response []map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "Should not error unmarshalling JSON")
	assert.Len(t, response, 1, "Should return one chat")

	firstUserExpectedUserId, err := utils.GetUserIDByUsername("newuser")
	if err != nil {
		t.Errorf("Error fetching user id: %v", err)
	}

	secondUserExpectedUserId, err := utils.GetUserIDByUsername("newuser2")
	if err != nil {
		t.Errorf("Error fetching user id: %v", err)
	}

	foundChat := false
	for _, chat := range response {
		users, ok := chat["users_involved_in_chat"].([]interface{})
		assert.True(t, ok, "users_involved_in_chat should be an array")
		assert.Contains(t, users, firstUserExpectedUserId)
		assert.Contains(t, users, secondUserExpectedUserId)
		if chat["name"] == "newuser/newuser2" {
			foundChat = true
		}
	}
	assert.True(t, foundChat, "Should find Chat 1")
}

func TestCheckChatExistsHandler(t *testing.T){
	secondUserExpectedUserId, err := utils.GetUserIDByUsername("newuser2")
	if err != nil {
		t.Errorf("Error fetching user id: %v", err)
	}

	requestData := map[string]interface{}{
		"user_id": secondUserExpectedUserId,
	}

	body, _ := json.Marshal(requestData)

	authHeader := fmt.Sprintf("Bearer %s", token)
	req := httptest.NewRequest("POST", "/chat-exists", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)

	w := httptest.NewRecorder()
	r := mux.NewRouter()
	protectedCheckChatExistsHandler := middleware.AuthMiddleware(http.HandlerFunc(checkChatExistsHandler))
	r.Handle("/chat-exists", protectedCheckChatExistsHandler)
	r.ServeHTTP(w, req)
	expectedChatId, err := GetChatIDByName("newuser/newuser2")
	if err != nil {
		t.Errorf("Error chat user id: %v", err)
	}

	assert.Equal(t, http.StatusOK, w.Code, "Status code should be OK")
	var responseMap map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &responseMap)
	fmt.Println(responseMap)
	assert.Equal(t, true, responseMap["exists"], "Should exist")
	assert.Equal(t, expectedChatId, responseMap["chat_id"], "Chat ID should match")
}