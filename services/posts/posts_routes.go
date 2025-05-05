package posts

import (
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"net/http"
	"github.com/Apower11/derivatives-pricing-microservice/middleware"
)

func RegisterRoutes(r *mux.Router) {	
	godotenv.Load()
	protectedCreatePostHandler := middleware.AuthMiddleware(http.HandlerFunc(createPostHandler))
	r.Handle("/post", protectedCreatePostHandler)
	protectedGetPostHandler := middleware.AuthMiddleware(http.HandlerFunc(getPostHandler))
	r.Handle("/post/{post_id}", protectedGetPostHandler)
	protectedGetPostsHandler := middleware.AuthMiddleware(http.HandlerFunc(getPostsHandler))
	r.Handle("/posts", protectedGetPostsHandler)
	r.HandleFunc("/posts/user", getPostsByUserIdHandler)
	r.HandleFunc("/posts/count", getNumberOfPostsHandler)
}