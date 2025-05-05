package auth

import(
	"net/http"
	"io"
	"github.com/Apower11/derivatives-pricing-microservice/utils"
	"github.com/Apower11/derivatives-pricing-microservice/db"
	"github.com/google/uuid"
	"fmt"
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"log"
	"gorm.io/gorm"
)

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	utils.EnableCORS(&w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading body: %v", err), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req UserRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	var count int64
	if err := db.DB.Raw("SELECT COUNT(*) FROM users WHERE username = ?", req.Username).Scan(&count).Error; err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		fmt.Println("DB error (count):", err)
		return
	}
	if count > 0 {
		http.Error(w, fmt.Sprintf("User '%s' exists", req.Username), http.StatusConflict)
		return
	}
	if req.Password != req.ConfirmPassword {
		http.Error(w, "Passwords mismatch", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, fmt.Sprintf("Hashing error: %v", err), http.StatusInternalServerError)
		log.Printf("Password hash error: %v", err)
		return
	}

	category := "customer"
	if req.IsAdmin {
		category = "admin"
	}

	newID := uuid.New().String()
	if err := db.DB.Exec("INSERT INTO users (id, username, password, current_balance, category) VALUES (?, ?, ?, 0, ?)", newID, req.Username, hashedPassword, category).Error; err != nil {
		log.Fatalf("DB insert error: %v", err)
		return
	}

	token, err := generateToken(newID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		fmt.Println("Token error:", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "User created", "token": token, "id": newID, "username": req.Username})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
        utils.EnableCORS(&w)
        if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var user User
	result := db.DB.Where("username = ?", req.Username).First(&user)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			fmt.Println("Database error:", result.Error)
		}
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := generateToken(user.ID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		fmt.Println("Token generation error:", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{Token: token, ID: user.ID, Username: user.Username, Category: user.Category})
}