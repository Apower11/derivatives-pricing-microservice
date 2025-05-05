package posts

import (
	"fmt"
	"net/http"
	"github.com/Apower11/derivatives-pricing-microservice/utils"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"github.com/Apower11/derivatives-pricing-microservice/db"
	"github.com/google/uuid"
	"encoding/json"
	"encoding/base64"
	"io"
	"strings"
	"time"
)

func createPostHandler(w http.ResponseWriter, r *http.Request) {
	utils.EnableCORS(&w)
	if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
	}
	
	if err := r.ParseMultipartForm(32 << 20); err != nil {
			http.Error(w, fmt.Sprintf("failed to parse multipart form: %v", err), http.StatusBadRequest)
			return
	}

	text := r.FormValue("text")
	if text == "" {
			http.Error(w, "text is required", http.StatusBadRequest)
			return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
			http.Error(w, fmt.Sprintf("failed to get image file: %v", err), http.StatusBadRequest)
			return
	}
	defer file.Close()

	imageData, err := io.ReadAll(file)
	if err != nil {
			http.Error(w, fmt.Sprintf("failed to read image data: %v", err), http.StatusInternalServerError)
			return
	}

	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		http.Error(w, "invalid file type: only images are allowed", http.StatusBadRequest)
		return
	}

	userIDStr := r.Context().Value("userID")
	now := time.Now()
	result := db.DB.Exec("INSERT INTO posts (id, user_id, text, image, created_at) VALUES ($1, $2, $3, $4, $5)", uuid.New(), userIDStr, text, imageData, now)
	if result.Error != nil {
		http.Error(w, fmt.Sprintf("failed to insert post into database: %v", result.Error), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "Post Created Successfully" }`,)
}

func getPostHandler(w http.ResponseWriter, r *http.Request) {
	utils.EnableCORS(&w)
	if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
	}
	vars := mux.Vars(r)
	postID := vars["post_id"]
	var post Post
	result := db.DB.Raw("SELECT id, user_id, text, image, created_at FROM posts WHERE id = ?", postID).Scan(&post)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			http.Error(w, "post not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("failed to query database: %v", result.Error), http.StatusInternalServerError)
		return
	}

	if post.Image == nil {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"id": "%s", "text": "%s", "created_at": "%s", "user_id": "%s" }`, post.ID, post.Text, post.CreatedAt.Format(time.RFC3339), post.UserID)
		return
	}

	imageType := getMIMEType(post.Image)
	w.Header().Set("Content-Type", "application/json")
	imageDataBase64 := base64.StdEncoding.EncodeToString(post.Image)
	fmt.Fprintf(w, `{"id": "%s", "text": "%s", "image": "data:%s;base64,%s", "created_at": "%s", "user_id": "%s" }`, post.ID, post.Text, imageType, imageDataBase64, post.CreatedAt.Format(time.RFC3339), post.UserID)
}

func getPostsHandler(w http.ResponseWriter, r *http.Request) {
	utils.EnableCORS(&w)
	if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
	}

	var posts []Post
	result := db.DB.Find(&posts)
	if result.Error != nil {
			http.Error(w, fmt.Sprintf("failed to query database: %v", result.Error), http.StatusInternalServerError)
			return
	}

	postsWithImages := make([]Post, len(posts))
	for i, post := range posts {
			postsWithImages[i] = post
			if post.Image != nil {
					imageType := getMIMEType(post.Image) // Get the MIME type.
					imageBase64 := base64.StdEncoding.EncodeToString(post.Image)
					postsWithImages[i].ImageBase64 = fmt.Sprintf("data:%s;base64,%s", imageType, imageBase64) //prepending
			}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(postsWithImages); err != nil {
			http.Error(w, fmt.Sprintf("failed to encode JSON: %v", err), http.StatusInternalServerError)
			return
	}
}

func getPostsByUserIdHandler(w http.ResponseWriter, r *http.Request) {
	utils.EnableCORS(&w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		http.Error(w, "user_id query parameter is required", http.StatusBadRequest)
		return
	}
	userID := userIDStr
	var posts []Post
	result := db.DB.Raw("SELECT id, user_id, text, image, created_at FROM posts WHERE user_id = ?", userID).Scan(&posts)
	if result.Error != nil {
		http.Error(w, fmt.Sprintf("failed to query database: %v", result.Error), http.StatusInternalServerError)
		return
	}

	postsWithImages := make([]Post, len(posts))
	for i, post := range posts {
			postsWithImages[i] = post
			if post.Image != nil {
					imageType := getMIMEType(post.Image)
					imageBase64 := base64.StdEncoding.EncodeToString(post.Image)
					postsWithImages[i].ImageBase64 = fmt.Sprintf("data:%s;base64,%s", imageType, imageBase64)
			}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(postsWithImages); err != nil {
			http.Error(w, fmt.Sprintf("failed to encode JSON: %v", err), http.StatusInternalServerError)
			return
	}
}

func getNumberOfPostsHandler(w http.ResponseWriter, r *http.Request) {
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
	result := db.DB.Raw("SELECT COUNT(*) FROM posts WHERE user_id = ?", userID).Scan(&count)
	if result.Error != nil {
			http.Error(w, fmt.Sprintf("failed to query database: %v", result.Error), http.StatusInternalServerError)
			return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"post_count": %d}`, count)
}