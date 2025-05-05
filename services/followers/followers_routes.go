package followers

import (
	"net/http"
	"github.com/gorilla/mux"
	"github.com/Apower11/derivatives-pricing-microservice/middleware"
)

func RegisterRoutes(r *mux.Router) {	
	protectedFollowRequestHandler := middleware.AuthMiddleware(http.HandlerFunc(followRequestHandler))
	r.Handle("/follow", protectedFollowRequestHandler)
	r.HandleFunc("/followers/count", getNumberOfFollowersHandler)
	r.HandleFunc("/followees/count", getNumberOfFolloweesHandler)
	protectedIsFollowingHandler := middleware.AuthMiddleware(http.HandlerFunc(isFollowingHandler))
	r.Handle("/following/check", protectedIsFollowingHandler)
}