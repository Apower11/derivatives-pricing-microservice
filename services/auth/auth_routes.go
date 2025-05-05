package auth

import (
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func RegisterRoutes(r *mux.Router) {	
	godotenv.Load()
	r.HandleFunc("/create-user", CreateUserHandler)
	r.HandleFunc("/login", loginHandler)
}