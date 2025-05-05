package portfolio

import (
	"net/http"
	"encoding/json"
	"github.com/Apower11/derivatives-pricing-microservice/utils"
	"github.com/Apower11/derivatives-pricing-microservice/types"
	"github.com/Apower11/derivatives-pricing-microservice/db"
	"fmt"
)

func addBalanceHandler(w http.ResponseWriter, r *http.Request) {
	utils.EnableCORS(&w)
	if r.Method == "OPTIONS" {
	w.WriteHeader(http.StatusOK)
	return
}
if r.Method != http.MethodPost {
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	return
}

var req AddBalanceRequest
err := json.NewDecoder(r.Body).Decode(&req)
if err != nil {
	http.Error(w, "Invalid request body", http.StatusBadRequest)
	return
}

userIDStr := r.Context().Value("userID")
if userIDStr == "" {
	http.Error(w, "User ID is required", http.StatusBadRequest)
	return
}

amountToAdd := req.Amount

result := db.DB.Exec("UPDATE users SET current_balance = current_balance + ? WHERE id = ?", amountToAdd, userIDStr)
if result.Error != nil {
	http.Error(w, "Internal server error", http.StatusInternalServerError)
	fmt.Println("Database update error:", result.Error)
	return
}

if result.RowsAffected == 0 {
	http.Error(w, "Failed to update user balance", http.StatusInternalServerError)
	return
}

var updatedUser types.User
if err := db.DB.First(&updatedUser, userIDStr).Error; err != nil {
	fmt.Println("Error fetching updated user:", err)
}

w.WriteHeader(http.StatusOK)
json.NewEncoder(w).Encode(map[string]interface{}{
	"message":         "Balance updated successfully",
	"new_balance":     updatedUser.CurrentBalance,
	"amount_added":    amountToAdd,
	"updated_user_id": userIDStr,
})
}

func getBalanceHandler(w http.ResponseWriter, r *http.Request) {
utils.EnableCORS(&w)
if r.Method == "OPTIONS" {
	w.WriteHeader(http.StatusOK)
	return
}

userIDFromContext := r.Context().Value("userID")
userIDStr, ok := userIDFromContext.(string) // Type assertion here
if !ok {
	http.Error(w, "User ID not found or not a string", http.StatusBadRequest)
	return
}
if userIDStr == "" {
	http.Error(w, "User ID is required", http.StatusBadRequest)
	return
}

var balance float64
query := "SELECT current_balance FROM users WHERE id = ?"
result := db.DB.Raw(query, userIDStr).Scan(&balance)
if result.Error != nil {
			http.Error(w, "failed to execute query", http.StatusInternalServerError)
	return 
}

if result.RowsAffected == 0 {
			http.Error(w, "Invalid user ID: "+userIDStr, http.StatusBadRequest)
	return
}

	response := map[string]float64{
	"balance": balance,
}

jsonResponse, err := json.Marshal(response)
if err != nil {
	http.Error(w, "Failed to marshal JSON: "+err.Error(), http.StatusInternalServerError)
	return
}

w.Header().Set("Content-Type", "application/json")
w.Write(jsonResponse)
}

func getUserAssetsHandler(w http.ResponseWriter, r *http.Request) {
	userIDCtx := r.Context().Value("userID")
	if userIDCtx == nil {
		http.Error(w, "Unauthorized access", http.StatusUnauthorized)
		return
	}

	userID, ok := userIDCtx.(string)
	if !ok {
		http.Error(w, "Invalid user ID in context", http.StatusInternalServerError)
		return
	}

	query := `
		SELECT 
			a.asset_id,
			a.name,
			SUM(CASE WHEN t.transaction_type = 'buy' THEN t.quantity ELSE -t.quantity END) as quantity,
                        a.symbol,
                        a.asset_type,
                        COALESCE(
				(SELECT AVG(p.price_per_unit) 
				 FROM transactions p
				 WHERE p.user_id = t.user_id
					 AND p.asset_id = a.asset_id
					 AND p.transaction_type = 'buy'
				), 0) as price_per_unit
		FROM transactions t
		JOIN assets a ON t.asset_id = a.asset_id
		WHERE t.user_id = ?
		GROUP BY a.asset_id, a.name, a.symbol, a.asset_type, t.user_id
		HAVING SUM(CASE WHEN t.transaction_type = 'buy' THEN t.quantity ELSE -t.quantity END) > 0
	`

	rows, err := db.DB.Raw(query, userID).Rows()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		fmt.Println("Error executing raw query:", err)
		return
	}
	defer rows.Close()

	var assets []AssetQuantity
	for rows.Next() {
		var asset AssetQuantity
		if err := rows.Scan(&asset.AssetID, &asset.AssetName, &asset.Quantity, &asset.Symbol, &asset.AssetType, &asset.PricePerUnit); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			fmt.Println("Error scanning row:", err)
			return
		}
		assets = append(assets, asset)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		fmt.Println("Error during row iteration:", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(assets)
}