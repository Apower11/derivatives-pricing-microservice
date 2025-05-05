package auth

import(
	"testing"
	"github.com/Apower11/derivatives-pricing-microservice/db"
	"github.com/Apower11/derivatives-pricing-microservice/types"
	"net/http"
	"net/http/httptest"
	"bytes"
	"errors"
	"gorm.io/gorm"
	"github.com/gorilla/mux"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"os"
	"time"
	"github.com/golang-jwt/jwt/v5"
)

func GetUserIDByUsername(username string) (string, error) {
    var user types.User
    result := db.DB.Where("username = ?", username).Select("id").First(&user)

    if result.Error != nil {
        if errors.Is(result.Error, gorm.ErrRecordNotFound) {
            return "", nil
        }
        return "", result.Error
    }
    return user.ID, nil
}

func TestCreateUserHandler(t *testing.T){
	db.InitTestDB()
	_ = db.DB.Exec("DELETE FROM users;")
	body := []byte(`{"username": "newuser", "password": "password123", "confirm_password": "password123", "is_admin": false}`)
	req := httptest.NewRequest("POST", "/create-user", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/create-user", CreateUserHandler)
	r.ServeHTTP(w, req)
	expectedUserId, err := GetUserIDByUsername("newuser")
	if err != nil {
		t.Errorf("Error fetching asset id: %v", err)
	}
	assert.Equal(t, http.StatusCreated, w.Code, "Status code should be Created")
	var responseMap map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &responseMap)
	assert.NoError(t, err, "Should not error unmarshalling JSON")
	assert.Equal(t, "User created", responseMap["message"], "Message should be 'User created'")
	assert.NotEmpty(t, responseMap["token"], "Token should not be empty")
	assert.Equal(t, expectedUserId, responseMap["id"], "IDs should match")
	assert.Equal(t, "newuser", responseMap["username"], "Username should match")
}

func TestLoginHandler(t *testing.T){
	body := []byte(`{"username": "newuser", "password": "password123" }`)
	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/login", loginHandler)
	r.ServeHTTP(w, req)
	expectedUserId, err := GetUserIDByUsername("newuser")
	if err != nil {
		t.Errorf("Error fetching asset id: %v", err)
	}
	assert.Equal(t, http.StatusOK, w.Code, "Status code should be OK")
	var responseMap map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &responseMap)
	assert.NoError(t, err, "Should not error unmarshalling JSON")
	assert.NotEmpty(t, responseMap["token"], "Token should not be empty")
	assert.Equal(t, expectedUserId, responseMap["id"], "IDs should match")
	assert.Equal(t, "newuser", responseMap["username"], "Username should match")
	assert.Equal(t, "customer", responseMap["category"], "Category should be customer")
}

func TestGenerateToken(t *testing.T) {
	testSecret := "test-secret-key"
	os.Setenv("jwtSecret", testSecret)
	defer os.Unsetenv("jwtSecret") 
	testUserID := "test-user-123"
	tokenString, err := generateToken(testUserID)
	assert.NoError(t, err, "generateToken should not return an error")
	assert.NotEmpty(t, tokenString, "Generated token should not be empty")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(testSecret), nil
	})
	assert.NoError(t, err, "Failed to parse the generated token")
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		assert.Equal(t, testUserID, claims["user_id"], "User ID in the token should match")
		expectedExp := time.Now().Add(time.Hour * 24).Unix()
		actualExp := int64(claims["exp"].(float64)) 
		assert.InDelta(t, expectedExp, actualExp, 2, "Expiration time should be approximately 24 hours from now")
		expectedNbf := time.Now().Unix()
		actualNbf := int64(claims["nbf"].(float64))
		assert.InDelta(t, expectedNbf, actualNbf, 2, "Not before time should be approximately now")
		expectedIat := time.Now().Unix()
		actualIat := int64(claims["iat"].(float64))
		assert.InDelta(t, expectedIat, actualIat, 2, "Issued at time should be approximately now")
	} else {
		t.Errorf("Token is invalid or claims are not valid")
	}
}