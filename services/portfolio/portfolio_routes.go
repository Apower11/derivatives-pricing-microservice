package portfolio

import (
	"net/http"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/Apower11/derivatives-pricing-microservice/middleware"
)

func RegisterRoutes(r *mux.Router) {	
	godotenv.Load()
	protectedAddBalanceHandler := middleware.AuthMiddleware(http.HandlerFunc(addBalanceHandler))
	r.Handle("/deposit", protectedAddBalanceHandler)
	protectedGetBalanceHandler := middleware.AuthMiddleware(http.HandlerFunc(getBalanceHandler))
	r.Handle("/current-balance", protectedGetBalanceHandler)
	protectedGetUserAssetsHandler := middleware.AuthMiddleware(http.HandlerFunc(getUserAssetsHandler))
	r.Handle("/user-assets", protectedGetUserAssetsHandler)
}