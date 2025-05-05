package derivatives

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"log"
	"io"
	"strings"
	"strconv"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/Apower11/derivatives-pricing-microservice/types"
	"github.com/Apower11/derivatives-pricing-microservice/utils"
	"github.com/Apower11/derivatives-pricing-microservice/db"
)

func addDerivativeHandler(w http.ResponseWriter, r *http.Request){
    utils.EnableCORS(&w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
			http.Error(w, "Error reading request body", http.StatusBadRequest)
			return
	}
	defer r.Body.Close()

	var data AddDerivativeRequestData
	err = json.Unmarshal(body, &data)
	if err != nil {
			http.Error(w, "Error decoding JSON", http.StatusBadRequest)
			return
	}
	derivativeID := uuid.New().String()
	assetID := uuid.New().String()

    sqlStatement := fmt.Sprintf("INSERT INTO derivatives_table VALUES ('%s', '%s', '%s', '%s', %f, CURRENT_TIMESTAMP, '%s', %f, ARRAY[]::PRICETIMESTAMP[], ARRAY[]::PRICETIMESTAMP[], ARRAY[]::PRICETIMESTAMP[], ARRAY[]::PRICETIMESTAMP[]);", derivativeID, data.Name, data.DerivativeType, data.AssetName, data.StrikePrice, data.ExpirationDate, data.RiskFreeRate)
	db.DB.Exec(sqlStatement)
	result := db.DB.Exec("INSERT INTO assets (asset_id, asset_type, symbol, name, derivative_id) VALUES (?, ?, ?, ?, ?)", assetID, "derivative", data.Symbol, data.Name, derivativeID)

	if result.Error != nil {
		http.Error(w, "Failed to add asset", http.StatusInternalServerError)
		return
	}

	jsonResponse := `{"Message": "Success"}`

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, jsonResponse)
}

func getHistoricalPrices(w http.ResponseWriter, r *http.Request) {
    utils.EnableCORS(&w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	mode := r.URL.Query().Get("mode")
	frequencyOfTimestamps := r.URL.Query().Get("frequency_of_timestamps")
    vars := mux.Vars(r)
	itemID := vars["derivative_id"]
	var timestamps string
    query := fmt.Sprintf("SELECT %s FROM derivatives_table WHERE id = '%s'", frequencyOfTimestamps, itemID)
	result := db.DB.Raw(query).Scan(&timestamps)
	if result.Error != nil {
		log.Fatalf("Failed to execute raw SQL query: %v", result.Error)
        return
	}

	timestamps = strings.ReplaceAll(timestamps, "\\", "")
	var timestampArray []types.PriceTimestamp
	if(timestamps == "{}"){
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("[]"))
			return
	}

	for _, value := range strings.Split(timestamps, `","`) {
		value = strings.NewReplacer(`"`, ``, `{`, ``, `}`, ``, `(`, ``, `)`, ``, `'`, ``).Replace(value)
		parts := strings.Split(value, ",")
		timestamp := parts[0]
		price, _ := strconv.ParseFloat(parts[1], 64)
		timestampArray = append(timestampArray, types.PriceTimestamp{TimestampOfPrice: timestamp, Price: price})
	}

    today := time.Now()
	var startDate time.Time
	switch mode {
		case "1D":
			startDate = today
		case "1W":
			startDate = today.AddDate(0, 0, -7)
		case "1M":
			startDate = today.AddDate(0, -1, 0)
		case "6M":
			startDate = today.AddDate(0, -6, 0)
		case "1Y":
			startDate = today.AddDate(-1, 0, 0)
		case "5Y":
			startDate = today.AddDate(-5, 0, 0)
		default:
			startDate = time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
	}
	startDateAtMidnight := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)

	filteredArray := []types.PriceTimestamp{}
	for _, item := range timestampArray {
		dateTime, err := time.ParseInLocation("2006-01-02 15:04:05", item.TimestampOfPrice, startDate.Location())
		if err != nil {
			continue
		}

		fmt.Println(item)
		fmt.Println(startDateAtMidnight)
		fmt.Println(time.Now())
		fmt.Println(dateTime.Before(time.Now()))

		if (dateTime.After(startDateAtMidnight) || dateTime.Equal(startDateAtMidnight)) && dateTime.Before(time.Now()) {
			filteredArray = append(filteredArray, item)
			fmt.Println(item)
		}
	}
    jsonData, err := json.Marshal(filteredArray)
	if err != nil {
		log.Fatalf("Error encoding to JSON: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func handleDerivative(w http.ResponseWriter, r *http.Request) {
    utils.EnableCORS(&w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
    vars := mux.Vars(r)
	itemID := vars["derivative_id"]
    assetID := r.URL.Query().Get("asset_id")
	derivativesJSON, err := getDerivativesJSON(string(itemID))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get derivatives: %v", err), http.StatusInternalServerError)
		return
	}

	now := time.Now()
	twentyFourHoursAgo := now.Add(-24 * time.Hour)
    rawQuery := fmt.Sprintf(`SELECT COALESCE(SUM(quantity), 0) as volume FROM transactions WHERE asset_id = ? AND transaction_date >= ?`)
	var assetVolume AssetVolume
	result := db.DB.Raw(rawQuery, assetID, twentyFourHoursAgo).Scan(&assetVolume)

	if result.Error != nil {
		http.Error(w, fmt.Sprintf("Failed to execute raw SQL query: %v", result.Error), http.StatusInternalServerError)
		return 
	}

	responseMap := map[string]interface{}{
		"derivative_data":  derivativesJSON, 
		"trading_volume": assetVolume.Volume,
	}

	jsonResponse, err := json.Marshal(responseMap)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal JSON: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(jsonResponse))
}

func handleDerivatives(w http.ResponseWriter, r *http.Request) {
    utils.EnableCORS(&w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	derivativesJSON, err := getDerivativesJSON("")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get derivatives: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(derivativesJSON))
}
