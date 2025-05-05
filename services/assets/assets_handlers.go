package assets

import (
	"net/http"
	"github.com/Apower11/derivatives-pricing-microservice/utils"
	"encoding/json"
	"github.com/Apower11/derivatives-pricing-microservice/db"
	"github.com/google/uuid"
	"fmt"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func addAssetHandler(w http.ResponseWriter, r *http.Request) {
	utils.EnableCORS(&w)
	if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AddAssetRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	allowedAssetTypes := map[string]bool{"stock": true, "cryptocurrency": true, "commodity": true, "derivative": true}
	if !allowedAssetTypes[req.AssetType] {
		http.Error(w, "Invalid asset_type", http.StatusBadRequest)
		return
	}

	result := db.DB.Exec("INSERT INTO assets (asset_id, asset_type, symbol, name) VALUES (?, ?, ?, ?)", uuid.New(), req.AssetType, req.Symbol, req.Name)
	if result.Error != nil {
		http.Error(w, "Failed to add asset", http.StatusInternalServerError)
		fmt.Println("Error creating asset:", result.Error)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Asset added successfully"})
}

func GetAssetsByTypeHandler(w http.ResponseWriter, r *http.Request) {
	utils.EnableCORS(&w)
	if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
	}
	assetType := r.URL.Query().Get("type")
		var assets []DBAsset
		var result *gorm.DB
	if assetType == "" {
				query := "SELECT * FROM assets"
				result = db.DB.Raw(query).Scan(&assets)
	} else {
				query := "SELECT * FROM assets WHERE asset_type = ?"
				result = db.DB.Raw(query, assetType).Scan(&assets)
		}

		if result.Error != nil {
				http.Error(w, fmt.Sprintf("Failed to fetch assets: %v", result.Error), http.StatusInternalServerError)
				return
		}

	if result.RowsAffected == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
		return
	}

	jsonResponse, err := json.Marshal(assets)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal JSON: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse) 
}

func GetAssetByIDHandler(w http.ResponseWriter, r *http.Request) {
	utils.EnableCORS(&w)
	if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
	}
	vars := mux.Vars(r)
	assetIDStr := vars["asset_id"]
	if assetIDStr == "" {
		http.Error(w, "Missing asset ID", http.StatusBadRequest)
		return
	}

	query := "SELECT * FROM assets WHERE asset_id = ?"
	var asset DBAsset
	result := db.DB.Raw(query, assetIDStr).Scan(&asset)
	if result.Error != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch asset: %v", result.Error), http.StatusInternalServerError)
		return
	}

	if result.RowsAffected == 0 {
		http.Error(w, fmt.Sprintf("Asset with ID %s not found", assetIDStr), http.StatusNotFound)
		return
	}

	jsonResponse, err := json.Marshal(asset)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal JSON: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}