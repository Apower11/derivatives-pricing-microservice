package transactions

import (
	"net/http"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/Apower11/derivatives-pricing-microservice/middleware"
)

func RegisterRoutes(r *mux.Router) {	
	godotenv.Load()
	protectedCreateBuyTransactionHandler := middleware.AuthMiddleware(http.HandlerFunc(createBuyTransactionHandler))
	r.Handle("/buy", protectedCreateBuyTransactionHandler)
	protectedCreateSellTransactionHandler := middleware.AuthMiddleware(http.HandlerFunc(createSellTransactionHandler))
	r.Handle("/sell", protectedCreateSellTransactionHandler)
}