package chats

import (
	"net/http"
	"github.com/gorilla/mux"
	"github.com/Apower11/derivatives-pricing-microservice/middleware"
)

func RegisterRoutes(r *mux.Router) {	
	protectedCreateChatHandler := middleware.AuthMiddleware(http.HandlerFunc(CreateChatHandler))
	r.Handle("/chat", protectedCreateChatHandler)
	protectedGetUserChatsHandler := middleware.AuthMiddleware(http.HandlerFunc(getUserChatsHandler))
	r.Handle("/chats", protectedGetUserChatsHandler)
	protectedCheckChatExistsHandler := middleware.AuthMiddleware(http.HandlerFunc(checkChatExistsHandler))
	r.Handle("/chat-exists", protectedCheckChatExistsHandler)
}