package followers

import(
	"testing"
	"github.com/Apower11/derivatives-pricing-microservice/services/auth"
	"github.com/Apower11/derivatives-pricing-microservice/utils"
	"github.com/Apower11/derivatives-pricing-microservice/middleware"
	"net/http"
	"net/http/httptest"
	"bytes"
	"fmt"
	"github.com/gorilla/mux"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/Apower11/derivatives-pricing-microservice/db"
)

var token string

func TestFollowRequestHandler(t *testing.T){
	db.InitTestDB()
	body := []byte(`{"username": "newuser", "password": "password123", "confirm_password": "password123", "is_admin": false}`)
	req := httptest.NewRequest("POST", "/create-user", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/create-user", auth.CreateUserHandler)
	r.ServeHTTP(w, req)

	var responseMap map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &responseMap)
	token = responseMap["token"]

	w = httptest.NewRecorder()
	body = []byte(`{"username": "newuser2", "password": "password123", "confirm_password": "password123", "is_admin": true}`)
	req = httptest.NewRequest("POST", "/create-user", bytes.NewBuffer(body))
	r.ServeHTTP(w, req)
	var secondUserExpectedUserId string
	secondUserExpectedUserId, err = utils.GetUserIDByUsername("newuser2")
	if err != nil {
		t.Errorf("Error fetching user id: %v", err)
	}

	protectedFollowRequestHandler := middleware.AuthMiddleware(http.HandlerFunc(followRequestHandler))
	r.Handle("/follow", protectedFollowRequestHandler)
	requestData := map[string]interface{}{
		"followee_user_id": secondUserExpectedUserId,
	}

	body, err = json.Marshal(requestData)
	authHeader := fmt.Sprintf("Bearer %s", token)
	req = httptest.NewRequest("POST", "/follow", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code, "Status code should be Created")
	expected := "{\"message\": \"You are now following the user.\" }" 
	if w.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %q want %q", w.Body.String(), expected)
	}
}

func TestGetNumberOfFollowersHandler(t *testing.T){
	w := httptest.NewRecorder()
	r := mux.NewRouter()

	var secondUserExpectedUserId string
	secondUserExpectedUserId, err := utils.GetUserIDByUsername("newuser2")
	if err != nil {
		t.Errorf("Error fetching user id: %v", err)
	}
	requestUrl := fmt.Sprintf("/followers/count?user_id=%s", secondUserExpectedUserId)

	r.HandleFunc("/followers/count", getNumberOfFollowersHandler)
	req := httptest.NewRequest("GET", requestUrl, nil)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Status code should be OK")
	expected := "{\"follower_count\": 1}" 
	if w.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %q want %q", w.Body.String(), expected)
	}
}

func TestGetNumberOfFolloweesHandler(t *testing.T){
	w := httptest.NewRecorder()
	r := mux.NewRouter()

	var firstUserExpectedUserId string
	firstUserExpectedUserId, err := utils.GetUserIDByUsername("newuser")
	if err != nil {
		t.Errorf("Error fetching user id: %v", err)
	}
	requestUrl := fmt.Sprintf("/followees/count?user_id=%s", firstUserExpectedUserId)

	r.HandleFunc("/followees/count", getNumberOfFolloweesHandler)
	req := httptest.NewRequest("GET", requestUrl, nil)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Status code should be OK")
	expected := "{\"following_count\": 1}" 
	if w.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %q want %q", w.Body.String(), expected)
	}
}

func TestIsFollowingHandler(t *testing.T){
	w := httptest.NewRecorder()
	r := mux.NewRouter()

	var secondUserExpectedUserId string
	secondUserExpectedUserId, err := utils.GetUserIDByUsername("newuser2")
	if err != nil {
		t.Errorf("Error fetching user id: %v", err)
	}

	protectedIsFollowingHandler := middleware.AuthMiddleware(http.HandlerFunc(isFollowingHandler))
	r.Handle("/following/check", protectedIsFollowingHandler)

	requestData := map[string]interface{}{
		"followee_user_id": secondUserExpectedUserId,
	}

	body, _ := json.Marshal(requestData)

	requestUrl := fmt.Sprintf("/following/check")
	authHeader := fmt.Sprintf("Bearer %s", token)
	req := httptest.NewRequest("POST", requestUrl, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Status code should be OK")
	expected := "{\"is_following\":true}\n" 
	if w.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %q want %q", w.Body.String(), expected)
	}
}