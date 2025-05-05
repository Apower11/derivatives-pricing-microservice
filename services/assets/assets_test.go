package assets

import(
	"testing"
	"github.com/Apower11/derivatives-pricing-microservice/db"
	"net/http"
	"net/http/httptest"
	"bytes"
	"errors"
	"gorm.io/gorm"
	"fmt"
	"github.com/gorilla/mux"
)

func GetAssetIDByName(name string) (string, error) {
    var asset Asset
    result := db.DB.Where("name = ?", name).Select("asset_id").First(&asset) //Select only the id

    if result.Error != nil {
        if errors.Is(result.Error, gorm.ErrRecordNotFound) {
            return "", nil // Or return a specific error like  return 0, gorm.ErrRecordNotFound
        }
        return "", result.Error
    }
    return asset.AssetID, nil
}

func TestAddAssetHandler(t *testing.T){
	db.InitTestDB()
	_ = db.DB.Exec("DELETE FROM assets;")
	body := []byte(`{"asset_type": "stock", "symbol": "AAPL", "name": "Apple Stock"}`)
	req, err := http.NewRequest(http.MethodPost, "/asset", bytes.NewBuffer(body))
	if err != nil {
		t.Errorf("could not create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json") // Set the content type

	rr := httptest.NewRecorder()
	addAssetHandler(rr, req)

	if status := rr.Code; status != 201 {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := "{\"message\":\"Asset added successfully\"}\n" // The order might vary
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %q want %q", rr.Body.String(), expected)
	}
}

func TestGetAssetsByTypeHandler(t *testing.T){
	req, err := http.NewRequest(http.MethodGet, "/assets?type=stock", nil)
	if err != nil {
		t.Errorf("could not create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json") // Set the content type

	rr := httptest.NewRecorder()
	GetAssetsByTypeHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expectedAssetId, err := GetAssetIDByName("Apple Stock")

	if err != nil {
		t.Errorf("Error fetching asset id: %v", err)
	}

	expected := fmt.Sprintf("[{\"AssetID\":\"%s\",\"Type\":\"stock\",\"DerivativeID\":\"\",\"Name\":\"Apple Stock\",\"Symbol\":\"AAPL\"}]", expectedAssetId)
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %q want %q", rr.Body.String(), expected)
	}
}

func TestGetAssetByIDHandler(t *testing.T){
	expectedAssetId, err := GetAssetIDByName("Apple Stock")
	if err != nil {
		t.Errorf("Error fetching asset id: %v", err)
	}
	requestURL := fmt.Sprintf("/assets/%s", expectedAssetId)
	fmt.Println(requestURL)
	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		t.Errorf("could not create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json") // Set the content type

	r := mux.NewRouter()
	r.HandleFunc("/assets/{asset_id}", GetAssetByIDHandler)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := fmt.Sprintf("{\"AssetID\":\"%s\",\"Type\":\"stock\",\"DerivativeID\":\"\",\"Name\":\"Apple Stock\",\"Symbol\":\"AAPL\"}", expectedAssetId)
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %q want %q", rr.Body.String(), expected)
	}
}