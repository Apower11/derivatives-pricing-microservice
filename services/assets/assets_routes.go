package assets

import (
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"net/http"
	"github.com/Apower11/derivatives-pricing-microservice/middleware"
)

func RegisterRoutes(r *mux.Router) {	
	godotenv.Load()
	protectedAddAssetHandler := middleware.AuthMiddleware(http.HandlerFunc(addAssetHandler))
	r.Handle("/asset", protectedAddAssetHandler)
	r.HandleFunc("/assets", GetAssetsByTypeHandler)
	r.HandleFunc("/assets/{asset_id}", GetAssetByIDHandler)
}