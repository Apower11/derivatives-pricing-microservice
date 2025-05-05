package messages

import (
	"net/http"
	"github.com/gorilla/mux"
	"github.com/Apower11/derivatives-pricing-microservice/middleware"
)

func RegisterRoutes(r *mux.Router) {	
	protectedAddMessageHandler := middleware.AuthMiddleware(http.HandlerFunc(addMessageHandler))
	r.Handle("/message", protectedAddMessageHandler)
	protectedGetMessagesByChatHandler := middleware.AuthMiddleware(http.HandlerFunc(getMessagesByChatHandler))
	r.Handle("/chats/{chat_id}/messages", protectedGetMessagesByChatHandler)
	r.HandleFunc("/ws/{chat_id}", handleWebSocket)
}