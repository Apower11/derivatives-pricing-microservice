package users

import (
	"fmt"
	"net/http"
	"github.com/Apower11/derivatives-pricing-microservice/utils"
	"github.com/Apower11/derivatives-pricing-microservice/types"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"github.com/Apower11/derivatives-pricing-microservice/db"
	"encoding/json"
	"strings"
)

func searchUsersHandler(w http.ResponseWriter, r *http.Request) {
	utils.EnableCORS(&w)
	if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
	}
	
	searchTerm := r.URL.Query().Get("search")
	if searchTerm == "" {
		http.Error(w, "search parameter is required", http.StatusBadRequest)
		return
	}

	searchTerm = strings.ReplaceAll(searchTerm, "%", "")
	var users []SearchUser
	result := db.DB.Raw("SELECT id, username FROM users WHERE username LIKE ?", searchTerm+"%").Scan(&users)
	if result.Error != nil {
		http.Error(w, fmt.Sprintf("failed to query database: %v", result.Error), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(users); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode JSON: %v", err), http.StatusInternalServerError)
		return
	}
}

func fetchAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	utils.EnableCORS(&w)
	if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
	}
	
	var users []SearchUser
	result := db.DB.Raw("SELECT id, username FROM users").Scan(&users)
	if result.Error != nil {
		http.Error(w, fmt.Sprintf("failed to query database: %v", result.Error), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(users); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode JSON: %v", err), http.StatusInternalServerError)
		return
	}
}

func getUsersByIdHandler(w http.ResponseWriter, r *http.Request) {
	utils.EnableCORS(&w)
	if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
	}
	
	vars := mux.Vars(r)
	userID := vars["user_id"]
	var user types.User
	result := db.DB.Raw("SELECT id, username FROM users WHERE id = ?", userID).Scan(&user)
	if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
					http.Error(w, "user not found", http.StatusNotFound)
					return
			}
			http.Error(w, fmt.Sprintf("failed to query database: %v", result.Error), http.StatusInternalServerError)
			return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, fmt.Sprintf("failed to encode JSON: %v", err), http.StatusInternalServerError)
			return
	}
}
