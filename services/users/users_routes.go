package users

import (
	"github.com/gorilla/mux"
)

func RegisterRoutes(r *mux.Router) {	
	r.HandleFunc("/users", searchUsersHandler)
	r.HandleFunc("/all-users", fetchAllUsersHandler)
	r.HandleFunc("/users/{user_id}", getUsersByIdHandler)
}