package followers

import (
	"fmt"
	"net/http"
	"github.com/Apower11/derivatives-pricing-microservice/utils"
	"github.com/Apower11/derivatives-pricing-microservice/db"
	"encoding/json"
	"io"
	"time"
)

func followRequestHandler(w http.ResponseWriter, r *http.Request) {
	utils.EnableCORS(&w)
	if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
	}
	followerUserID := r.Context().Value("userID") 
	body, err := io.ReadAll(r.Body)
	if err != nil {
			http.Error(w, fmt.Sprintf("failed to read request body: %v", err), http.StatusBadRequest)
			return
	}
	defer r.Body.Close()

	var followRequest FollowRequest
	if err := json.Unmarshal(body, &followRequest); err != nil {
			http.Error(w, fmt.Sprintf("failed to unmarshal JSON: %v", err), http.StatusBadRequest)
			return
	}

	if followRequest.FolloweeUserID == "" {
			http.Error(w, "invalid followee_user_id", http.StatusBadRequest)
			return
	}

	if followerUserID == followRequest.FolloweeUserID {
			http.Error(w, "cannot follow yourself", http.StatusBadRequest)
			return
	}

	now := time.Now()
	result := db.DB.Exec(
			"INSERT INTO followers (follower_user_id, followee_user_id, created_at) VALUES ($1, $2, $3)",
			followerUserID, followRequest.FolloweeUserID, now,
	)
	if result.Error != nil {
			http.Error(w, fmt.Sprintf("failed to insert follower: %v", result.Error), http.StatusInternalServerError)
			return
	}

	if result.RowsAffected != 1 {
			http.Error(w, "failed to create follower relationship: no rows inserted", http.StatusInternalServerError)
			return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "You are now following the user." }`,)
}

func getNumberOfFollowersHandler(w http.ResponseWriter, r *http.Request) {
	utils.EnableCORS(&w)
	if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
	}
	
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
			http.Error(w, "user_id query parameter is required", http.StatusBadRequest)
			return
	}

	var count int64
	result := db.DB.Raw("SELECT COUNT(*) FROM followers WHERE followee_user_id = ?", userID).Scan(&count)
	if result.Error != nil {
			http.Error(w, fmt.Sprintf("failed to query database: %v", result.Error), http.StatusInternalServerError)
			return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"follower_count": %d}`, count)
}

func getNumberOfFolloweesHandler(w http.ResponseWriter, r *http.Request) {
	utils.EnableCORS(&w)
	if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
	}
	
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
			http.Error(w, "user_id query parameter is required", http.StatusBadRequest)
			return
	}

	var count int64
	result := db.DB.Raw("SELECT COUNT(*) FROM followers WHERE follower_user_id = ?", userID).Scan(&count)
	if result.Error != nil {
			http.Error(w, fmt.Sprintf("failed to query database: %v", result.Error), http.StatusInternalServerError)
			return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"following_count": %d}`, count)
}

func isFollowingHandler(w http.ResponseWriter, r *http.Request) {
	utils.EnableCORS(&w)
	if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
	}
	
	followerUserID := r.Context().Value("userID")
	body, err := io.ReadAll(r.Body)
	if err != nil {
			http.Error(w, fmt.Sprintf("failed to read request body: %v", err), http.StatusBadRequest)
			return
	}
	defer r.Body.Close()

	var followCheckRequest FollowCheckRequest
	if err := json.Unmarshal(body, &followCheckRequest); err != nil {
			http.Error(w, fmt.Sprintf("failed to unmarshal JSON: %v", err), http.StatusBadRequest)
			return
	}

	if followCheckRequest.FolloweeUserID == "" {
			http.Error(w, "invalid followee_user_id", http.StatusBadRequest)
			return
	}

	var count int64
	result := db.DB.Raw("SELECT COUNT(*) FROM followers WHERE follower_user_id = ? AND followee_user_id = ?", followerUserID, followCheckRequest.FolloweeUserID).Scan(&count)
	if result.Error != nil {
			http.Error(w, fmt.Sprintf("failed to query database: %v", result.Error), http.StatusInternalServerError)
			return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]bool{
			"is_following": count > 0,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, fmt.Sprintf("failed to encode JSON: %v", err), http.StatusInternalServerError)
			return
	}
}