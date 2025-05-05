package derivatives

import (
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func RegisterRoutes(r *mux.Router) {	
	godotenv.Load()
	r.HandleFunc("/derivatives", handleDerivatives)
	r.HandleFunc("/derivative/{derivative_id}", handleDerivative)
	r.HandleFunc("/derivatives/historical-data/{derivative_id}", getHistoricalPrices)
	r.HandleFunc("/add-derivative", addDerivativeHandler)
}