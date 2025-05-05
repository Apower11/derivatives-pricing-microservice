package transactions

import(
	"net/http"
	"encoding/json"
	"github.com/Apower11/derivatives-pricing-microservice/db"
	"gorm.io/gorm"
	"github.com/Apower11/derivatives-pricing-microservice/types"
	"fmt"
	"errors"
	"github.com/google/uuid"
	"time"
)

func createBuyTransactionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	userIDCtx := r.Context().Value("userID")
	userID, ok := userIDCtx.(string)
	if !ok || userIDCtx == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	var req CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	
	var asset types.Asset
	if err := db.DB.First(&asset, "asset_id = ?", req.AssetID).Error; err != nil {
		status := http.StatusInternalServerError
		msg := "Internal server error"
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = http.StatusNotFound
			msg = "Asset not found"
		}
		http.Error(w, msg, status)
		fmt.Println("DB error (asset):", err)
		return
	}
	
	var user types.User
	if err := db.DB.First(&user, "id = ?", userID).Error; err != nil {
		status := http.StatusInternalServerError
		msg := "Internal server error"
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = http.StatusNotFound
			msg = "User not found"
		}
		http.Error(w, msg, status)
		fmt.Println("DB error (user):", err)
		return
	}
	
	cost := req.Quantity * req.PricePerUnit
	if user.CurrentBalance < cost {
		http.Error(w, "Insufficient funds", http.StatusBadRequest)
		return
	}
	
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&user).Where("id = ?", userID).Update("current_balance", user.CurrentBalance-cost).Error; err != nil {
			return err
		}
		newTx := Transaction{
			TransactionID:   uuid.New().String(),
			UserID:          userID,
			AssetID:         req.AssetID,
			TransactionType: "buy",
			Quantity:        req.Quantity,
			PricePerUnit:    req.PricePerUnit,
			TransactionDate: time.Now(),
		}
		return tx.Create(&newTx).Error
	})
	
	if err != nil {
		http.Error(w, "Transaction failed", http.StatusInternalServerError)
		fmt.Println("Transaction error:", err)
		return
	}
	
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Transaction created", "transaction_id": uuid.New().String()})
}

func createSellTransactionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	userIDCtx := r.Context().Value("userID")
	userID, ok := userIDCtx.(string)
	if !ok || userIDCtx == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	var req CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	
	var asset types.Asset
	if err := db.DB.First(&asset, "asset_id = ?", req.AssetID).Error; err != nil {
		status := http.StatusInternalServerError
		msg := "Internal server error"
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = http.StatusNotFound
			msg = "Asset not found"
		}
		http.Error(w, msg, status)
		fmt.Println("DB error (asset):", err)
		return
	}
	
	var user types.User
	if err := db.DB.First(&user, "id = ?", userID).Error; err != nil {
		status := http.StatusInternalServerError
		msg := "Internal server error"
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = http.StatusNotFound
			msg = "User not found"
		}
		http.Error(w, msg, status)
		fmt.Println("DB error (user):", err)
		return
	}
	
	saleAmount := req.Quantity * req.PricePerUnit
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&user).Where("id = ?", userID).Update("current_balance", user.CurrentBalance+saleAmount).Error; err != nil {
			return err
		}
		newTx := Transaction{
			TransactionID:   uuid.New().String(),
			UserID:          userID,
			AssetID:         req.AssetID,
			TransactionType: "sell",
			Quantity:        req.Quantity,
			PricePerUnit:    req.PricePerUnit,
			TransactionDate: time.Now(),
		}
		return tx.Create(&newTx).Error
	})
	
	if err != nil {
		http.Error(w, "Transaction failed", http.StatusInternalServerError)
		fmt.Println("Transaction error:", err)
		return
	}
	
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Sell transaction created", "transaction_id": uuid.New().String()})
}