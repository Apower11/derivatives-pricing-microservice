package derivatives

import (
	"testing"
	"github.com/Apower11/derivatives-pricing-microservice/cache"
	"github.com/Apower11/derivatives-pricing-microservice/utils"
	"github.com/Apower11/derivatives-pricing-microservice/db"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"io"
	"time"
	"strings"
	"bytes"
	"errors"
	"gorm.io/gorm"
	"encoding/json"
)

func GetDerivativeIDByName(name string) (string, error) {
    var derivative DBDerivative
	query := fmt.Sprintf("SELECT * FROM %s WHERE derivative_name = '%s';", "derivatives_table", name)
    result := db.DB.Raw(query).Scan(&derivative)
	if result.Error != nil {
		return "", fmt.Errorf("Failed to execute raw SQL query: %v", result.Error)
	}

    if result.Error != nil {
        if errors.Is(result.Error, gorm.ErrRecordNotFound) {
            return "", nil
        }
        return "", result.Error
    }
    return derivative.DerivativeID, nil
}

func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

type RoundTripFunc func(*http.Request) (*http.Response, error)
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestFetchPrices(t *testing.T){
	os.Setenv("apiKey", "b8mf8ejRIA2IuYEdsiiyZw==8gH9zNbHhCjK6rki")
	defer os.Unsetenv("apiKey")

	mockHTTPClient := NewTestClient(func(req *http.Request) (*http.Response, error) {
		assert.Contains(t, req.URL.String(), "api-ninjas.com", "Request should be to the mock API")
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`[{"close": 10.50}, {"close": 11.25}, {"close": 12.00}]`)),
			Header: map[string][]string{"Content-Type": {"application/json"}},
			Request:    req,
		}, nil
	})

	originalHTTPClient := http.DefaultClient
	http.DefaultClient = mockHTTPClient
	defer func() {
		http.DefaultClient = originalHTTPClient // Restore original client.
	}()

	startTimestamp := time.Now().Add(-24 * time.Hour).Unix()
	endTimestamp := time.Now().Unix()
	prices, err := fetchPrices(startTimestamp, endTimestamp, "any_name")

	assert.NoError(t, err, "Expected no error")
	assert.Equal(t, []float64{10.50, 11.25, 12.00}, prices, "Prices should match the mock response")
}

func TestFetchCurrentPrice(t *testing.T){
	os.Setenv("apiKey", "b8mf8ejRIA2IuYEdsiiyZw==8gH9zNbHhCjK6rki")
	defer os.Unsetenv("apiKey")

	mockHTTPClient := NewTestClient(func(req *http.Request) (*http.Response, error) {
		assert.Contains(t, req.URL.String(), "api-ninjas.com", "Request should be to the mock API")
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"price": 123.45}`)),
			Header:       map[string][]string{"Content-Type": {"application/json"}},
			Request:    req,
		}, nil
	})

	originalHTTPClient := http.DefaultClient
	http.DefaultClient = mockHTTPClient
	defer func() {
		http.DefaultClient = originalHTTPClient
	}()

	price, err := fetchCurrentPrice("some_commodity")

	assert.NoError(t, err, "Expected no error")
	assert.Equal(t, 123.45, price, "Prices should match the mock response")
}

func TestBlackScholes(t *testing.T) {
	derivative1 := DBDerivative{
		DerivativeType: "call",
		SpotPrice:      100.0,
		StrikePrice:    100.0,
		TimeToExpiry:   1.0,
		RiskFreeRate:   0.05,
		Volatility:     0.20,
	}
	expectedCallPrice1 := 10.450583572185565
	actualCallPrice1, err := BlackScholes(derivative1)
	assert.NoError(t, err)
	assert.InDelta(t, expectedCallPrice1, actualCallPrice1, 0.000001, "Test Case 1: Call Option")

	derivative2 := DBDerivative{
		DerivativeType: "put",
		SpotPrice:      100.0,
		StrikePrice:    100.0,
		TimeToExpiry:   1.0,
		RiskFreeRate:   0.05,
		Volatility:     0.20,
	}
	expectedPutPrice2 := 5.573526022256971
	actualPutPrice2, err := BlackScholes(derivative2)
	assert.NoError(t, err)
	assert.InDelta(t, expectedPutPrice2, actualPutPrice2, 0.000001, "Test Case 2: Put Option")

	derivative3 := DBDerivative{
		DerivativeType: "call",
		SpotPrice:      100.0,
		StrikePrice:    100.0,
		TimeToExpiry:   0,
		RiskFreeRate:   0.05,
		Volatility:     0.20,
	}
	_, err3 := BlackScholes(derivative3)
	assert.Error(t, err3, "Test Case 3: Time to expiry is 0")

	derivative4 := DBDerivative{
		DerivativeType: "put",
		SpotPrice:      0,
		StrikePrice:    100.0,
		TimeToExpiry:   1.0,
		RiskFreeRate:   0.05,
		Volatility:     0.20,
	}
	_, err4 := BlackScholes(derivative4)
	assert.Error(t, err4, "Test Case 4: Spot price is 0")

	derivative5 := DBDerivative{
		DerivativeType: "invalid",
		SpotPrice:      100.0,
		StrikePrice:    100.0,
		TimeToExpiry:   1.0,
		RiskFreeRate:   0.05,
		Volatility:     0.20,
	}
	_, err5 := BlackScholes(derivative5)
	assert.Error(t, err5, "Test Case 5: Invalid option type")

    derivative6 := DBDerivative{
        DerivativeType: "call",
        SpotPrice:      100.0,
        StrikePrice:    100.0,
        TimeToExpiry:   1.0,
        RiskFreeRate:   0.05,
        Volatility:     0.0,
    }
    _, err6 := BlackScholes(derivative6)
    assert.Error(t, err6, "Test Case 6: Zero volatility")
}

func TestFutureOrForwardPrice(t *testing.T) {
	derivative1 := DBDerivative{
		SpotPrice:    100.0,
		RiskFreeRate: 0.05,
		TimeToExpiry: 1.0,
	}
	expectedPrice1 := 105.12710963760241
	actualPrice1 := FutureOrForwardPrice(derivative1)
	assert.InDelta(t, expectedPrice1, actualPrice1, 0.000001, "Test Case 1: Standard Values")

	derivative2 := DBDerivative{
		SpotPrice:    100.0,
		RiskFreeRate: 0.05,
		TimeToExpiry: 0.0,
	}
	expectedPrice2 := 100.0
	actualPrice2 := FutureOrForwardPrice(derivative2)
	assert.InDelta(t, expectedPrice2, actualPrice2, 0.000001, "Test Case 2: Time to Expiry is 0")

	derivative3 := DBDerivative{
		SpotPrice:    100.0,
		RiskFreeRate: 0.0,
		TimeToExpiry: 1.0,
	}
	expectedPrice3 := 100.0
	actualPrice3 := FutureOrForwardPrice(derivative3)
	assert.InDelta(t, expectedPrice3, actualPrice3, 0.000001, "Test Case 3: Risk-free Rate is 0")
}

func TestCalculateVolatility(t *testing.T) {
	prices1 := []float64{100.0, 101.0, 102.0, 103.0, 102.0, 104.0}
	daysInYear1 := 252
	expectedVolatility1 := 0.15212681856638915
	actualVolatility1, err1 := CalculateVolatility(prices1, daysInYear1)
	assert.NoError(t, err1)
	assert.InDelta(t, expectedVolatility1, actualVolatility1, 0.000001, "Test Case 1: Standard Values")

	prices2 := []float64{100.0, 105.0}
	daysInYear2 := 252
	expectedVolatility2 := 0
	actualVolatility2, err2 := CalculateVolatility(prices2, daysInYear2)
	assert.NoError(t, err2)
	assert.InDelta(t, expectedVolatility2, actualVolatility2, 0.000001, "Test Case 2: Two Prices")

	prices3 := []float64{}
	daysInYear3 := 252
	_, err3 := CalculateVolatility(prices3, daysInYear3)
	assert.Error(t, err3, "Test Case 3: Empty Prices Array")

    prices4 := []float64{100.0}
    daysInYear4 := 252
    _, err4 := CalculateVolatility(prices4, daysInYear4)
    assert.Error(t, err4, "Test Case 4: Single Price")
}

func TestCalculateTimeToExpiry(t *testing.T) {
	expirationDate1 := time.Now().AddDate(1, 0, 0)
	expectedExpiry1 := 1.0
	actualExpiry1 := calculateTimeToExpiry(expirationDate1)
	assert.InDelta(t, expectedExpiry1, actualExpiry1, 0.001, "Test Case 1: 1 Year Expiry")

	expirationDate2 := time.Now().AddDate(0, 6, 0)
	expectedExpiry2 := 0.5038786219484371
	actualExpiry2 := calculateTimeToExpiry(expirationDate2)
	assert.InDelta(t, expectedExpiry2, actualExpiry2, 0.001, "Test Case 2: 6 Months Expiry")

	expirationDate3 := time.Now()
	expectedExpiry3 := 0.0
	actualExpiry3 := calculateTimeToExpiry(expirationDate3)
	assert.InDelta(t, expectedExpiry3, actualExpiry3, 0.001, "Test Case 3: Expiration Today")
}

func TestPriceDerivative(t *testing.T) {
	os.Setenv("apiKey", "b8mf8ejRIA2IuYEdsiiyZw==8gH9zNbHhCjK6rki")
	derivative1 := DBDerivative{
		DerivativeType: "call",
		SpotPrice:      100.0,
		StrikePrice:    100.0,
		TimeToExpiry:   1.0,
		RiskFreeRate:   0.05,
		AssetName:      "silver", 
		DerivativeName: "Example",
	}
	thisYearTimestamp, _ := utils.GetStartOfTodayTimestamps()
	volatilityCacheKey := fmt.Sprintf("%s:%s", strings.ToLower("Example"), fmt.Sprint(thisYearTimestamp))
	now := time.Now()
	tomorrow := now.AddDate(0, 0, 1)
	midnight := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, now.Location())
	cache.CacheItem(volatilityCacheKey, 0.2, midnight.Sub(now))

	expectedPrice1 := 10.45
	actualPrice1, err := priceDerivative(derivative1)
	assert.NoError(t, err)
	assert.InDelta(t, expectedPrice1, actualPrice1, 0.01, "Test Case 1: Call Option, no cache")

	derivative2 := DBDerivative{
		DerivativeType: "put",
		SpotPrice:      100.0,
		StrikePrice:    100.0,
		TimeToExpiry:   1.0,
		RiskFreeRate:   0.05,
		AssetName:      "silver",
		DerivativeName: "Example", 
	}

	expectedPrice2 := 5.57
	actualPrice2, err := priceDerivative(derivative2)
	assert.NoError(t, err)
	assert.InDelta(t, expectedPrice2, actualPrice2, 0.01, "Test Case 2: Put Option, cache hit")

	derivative3 := DBDerivative{
		DerivativeType: "future",
		SpotPrice:      100.0,
		RiskFreeRate:   0.05,
		TimeToExpiry:   1.0,
		AssetName:      "silver",
		DerivativeName: "Example",
	}
	expectedPrice3 := 105.13
	actualPrice3, err := priceDerivative(derivative3)
	assert.NoError(t, err)
	assert.InDelta(t, expectedPrice3, actualPrice3, 0.01, "Test Case 3: Future")
}

func TestAddDerivativeHandler(t *testing.T){
	db.InitTestDB()
	_ = db.DB.Exec("DELETE FROM derivatives_table;")
	body := []byte(`{"name": "Gold Future", "assetName": "gold", "derivativeType": "future", "strikePrice": 4000, "expirationDate": "2027-04-17 15:05:48.34612", "riskFreeRate": 0.03, "symbol": "gold.future.1"}`)
	req, err := http.NewRequest(http.MethodPost, "/asset", bytes.NewBuffer(body))
	if err != nil {
		t.Errorf("could not create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json") 

	rr := httptest.NewRecorder()
	addDerivativeHandler(rr, req)

	if status := rr.Code; status != 200 {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := "{\"Message\": \"Success\"}" 
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %q want %q", rr.Body.String(), expected)
	}
}

func TestGetDerivativesJSON(t *testing.T) {
	expectedDerivativeId, err := GetDerivativeIDByName("Gold Future")

	jsonResult, err := getDerivativesJSON("")
	assert.NoError(t, err)

	var returnedDerivatives []DBDerivative
	err = json.Unmarshal([]byte(jsonResult), &returnedDerivatives)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(returnedDerivatives), "Should return 1 derivatives")
	assert.Equal(t, "Gold Future", returnedDerivatives[0].DerivativeName, "Asset name should match")
	assert.Equal(t, expectedDerivativeId, returnedDerivatives[0].DerivativeID, "Asset name should match")
	assert.Equal(t, "future", returnedDerivatives[0].DerivativeType, "Asset name should match")
	assert.Equal(t, "gold", returnedDerivatives[0].AssetName, "Asset name should match")
	assert.Equal(t, 4000.00, returnedDerivatives[0].StrikePrice, "Asset name should match")
	assert.Equal(t, 0.03, returnedDerivatives[0].RiskFreeRate, "Asset name should match")
	assert.Equal(t, time.Time(time.Date(2027, time.April, 17, 15, 5, 48, 346120000, time.UTC)), returnedDerivatives[0].ExpirationDate, "Asset name should match")
}