package derivatives

import (
	"log"
	"time"
	"fmt"
	"net/http"
	"io"
	"os"
	"encoding/json"
	"math"
	"strings"
	"github.com/Apower11/derivatives-pricing-microservice/utils"
	"github.com/Apower11/derivatives-pricing-microservice/db"
	"github.com/Apower11/derivatives-pricing-microservice/cache"
)

func RunEvery(d time.Duration, fn func()) {
	now := time.Now()
	next := now.Truncate(d).Add(d)
	if next.Before(now) {
		next = next.Add(d)
	}
	initialDelay := next.Sub(now)
	time.AfterFunc(initialDelay, func() {
		ticker := time.NewTicker(d)
		defer ticker.Stop()
		for range ticker.C {
			fn()
		}
	})
}

func RunEveryFifthMinuteFromBeginningOfDay(fn func()) {
	RunEvery(5*time.Minute, fn)
}

func RunEveryHalfHour(fn func()) {
	RunEvery(30*time.Minute, fn)
}

func RunAtMidnight(fn func()) {
	RunEvery(24*time.Hour, fn)
}

func RunAtStartOfWeek(fn func()) {
	now := time.Now()
	loc := now.Location()
	daysUntilMonday := (time.Monday - now.Weekday() + 7) % 7
	nextMondayMidnight := time.Date(now.Year(), now.Month(), now.Day()+int(daysUntilMonday), 0, 0, 0, 0, loc)
	if now.Weekday() == time.Monday && now.Hour() >= 0 {
		nextMondayMidnight = nextMondayMidnight.Add(7 * 24 * time.Hour)
	}
	initialDelay := nextMondayMidnight.Sub(now)
	time.AfterFunc(initialDelay, func() {
		ticker := time.NewTicker(7 * 24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			fn()
		}
	})
}

func fetchPrices(startTimestamp, endTimestamp int64, name string) ([]float64, error) {
	apiUrl := fmt.Sprintf("https://api.api-ninjas.com/v1/commoditypricehistorical?name=%s&period=1d&start=%d&end=%d", name, startTimestamp, endTimestamp)
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("X-API-KEY", os.Getenv("apiKey"))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading body: %w", err)
	}

	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		return nil, fmt.Errorf("API error %s: %s", resp.Status, body)
	}

	var priceObjects []struct {
		Close float64 `json:"close"`
	}
	if err := json.Unmarshal(body, &priceObjects); err != nil {
		return nil, fmt.Errorf("error unmarshalling: %w. Body: %s", err, body)
	}

	prices := make([]float64, len(priceObjects))
	for index, price := range priceObjects {
		prices[index] = price.Close
	}

	if len(prices) == 0 {
		return nil, fmt.Errorf("no prices found")
	}

	return prices, nil
}

func fetchCurrentPrice(name string) (float64, error) {
	apiUrl := fmt.Sprintf("https://api.api-ninjas.com/v1/commodityprice?name=%s", name)

	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		return 0, fmt.Errorf("error creating HTTP request: %w", err)
	}

	req.Header.Add("X-API-KEY", os.Getenv("apiKey"))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("error making HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, fmt.Errorf("API request failed with status %s, body: %s", resp.Status, string(body))
	}

	var response struct {
		Price float64 `json:"price"`
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0, fmt.Errorf("error unmarshalling JSON: %w, body: %s", err, string(body))
	}

	return response.Price, nil
}

func BlackScholes(derivative DBDerivative) (float64, error) {
        S := derivative.SpotPrice
        K := derivative.StrikePrice
        T := derivative.TimeToExpiry
        r := derivative.RiskFreeRate
        sigma := derivative.Volatility

        if S <= 0 || K <= 0 || T <= 0 || sigma <= 0 {
                return 0, fmt.Errorf("invalid input parameters")
        }

        d1 := (math.Log(S/K) + (r+0.5*sigma*sigma)*T) / (sigma * math.Sqrt(T))
        d2 := d1 - sigma*math.Sqrt(T)

        Nd1 := 0.5 * (1 + math.Erf(d1/math.Sqrt(2)))
        Nd2 := 0.5 * (1 + math.Erf(d2/math.Sqrt(2)))
        Nd1Neg := 0.5 * (1 + math.Erf(-d1/math.Sqrt(2)))
        Nd2Neg := 0.5 * (1 + math.Erf(-d2/math.Sqrt(2)))

        switch strings.ToLower(derivative.DerivativeType) {
        case "call":
                return S*Nd1 - K*math.Exp(-r*T)*Nd2, nil
        case "put":
                return K*math.Exp(-r*T)*Nd2Neg - S*Nd1Neg, nil
        default:
                return 0, fmt.Errorf("invalid derivative type")
        }
}

func FutureOrForwardPrice(derivative DBDerivative) float64 {
        S := derivative.SpotPrice
        r := derivative.RiskFreeRate
        T := derivative.TimeToExpiry
        return S * math.Exp(r*T)
}

func CalculateVolatility(prices []float64, daysInYear int) (float64, error) {
	if len(prices) < 2 {
		return 0, fmt.Errorf("need at least 2 data points to calculate volatility")
	}

	returns := make([]float64, len(prices)-1)
	for i := 1; i < len(prices); i++ {
		returns[i-1] = (prices[i] - prices[i-1]) / prices[i-1]
	}

	meanReturn := 0.0
	for _, ret := range returns {
		meanReturn += ret
	}
	meanReturn /= float64(len(returns))

	sumSquaredDeviations := 0.0
	for _, ret := range returns {
		deviation := ret - meanReturn
		sumSquaredDeviations += deviation * deviation
	}

	variance := sumSquaredDeviations / float64(len(returns))
	dailyVolatility := math.Sqrt(variance)
	annualizedVolatility := dailyVolatility * math.Sqrt(float64(daysInYear))

	return annualizedVolatility, nil
}

func calculateTimeToExpiry(expirationDate time.Time) float64 {
	now := time.Now()
	duration := expirationDate.Sub(now)
	return duration.Hours() / (24.0 * 365.25)
}

func priceDerivative(derivative DBDerivative) (float64, error) {
	thisYearTimestamp, lastYearTimestamp := utils.GetStartOfTodayTimestamps()
	volatilityCacheKey := fmt.Sprintf("%s:%s", strings.ToLower(derivative.DerivativeName), fmt.Sprint(thisYearTimestamp))
	val, found := cache.IsCached(volatilityCacheKey)
	if found {
			derivative.Volatility = val.(float64)
	} else {
			prices, err := fetchPrices(lastYearTimestamp, thisYearTimestamp, derivative.AssetName)
			if err != nil {
					return 0, fmt.Errorf("Error fetching prices for volatility calculation: %v", err)
			}
			volatility, err := CalculateVolatility(prices, 365)
			if err != nil {
				return 0, fmt.Errorf("Error calculating volatility: %v", err)
			}
			now := time.Now()
			tomorrow := now.AddDate(0, 0, 1)
			midnight := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, now.Location())
			cache.CacheItem(volatilityCacheKey, volatility, midnight.Sub(now))
			derivative.Volatility = volatility
	}
	
	switch strings.ToLower(derivative.DerivativeType) {
	case "call", "put":
			derivativePrice, err := BlackScholes(derivative)
			if err != nil {
				return 0, fmt.Errorf("Error pricing option: %v", err)
			}
			derivative.CurrentPrice = utils.RoundFloat(derivativePrice, 2)
	case "future", "forward":
			derivativePrice := FutureOrForwardPrice(derivative)
			derivative.CurrentPrice = utils.RoundFloat(derivativePrice, 2)
	}
	return derivative.CurrentPrice, nil
}

func getDerivativesJSON(itemID string) (string, error) {
	var derivatives []DBDerivative
    var err error
	var query string
	if(itemID == ""){
		query = fmt.Sprintf("SELECT * FROM derivatives_table;")
	} else {
		query = fmt.Sprintf("SELECT * FROM derivatives_table WHERE id = '%s';", itemID)
	}
	result := db.DB.Raw(query).Scan(&derivatives)
	if result.Error != nil {
		log.Fatalf("Failed to execute raw SQL query: %v", result.Error)
	}

	for i := range derivatives {
		derivatives[i].TimeToExpiry = calculateTimeToExpiry(derivatives[i].ExpirationDate)
		derivatives[i].SpotPrice, err = fetchCurrentPrice(strings.ToLower(derivatives[i].AssetName))
		if err != nil {
			return "", fmt.Errorf("Error fetching Spot Price: %v", err)
		}
		derivatives[i].CurrentPrice, err = priceDerivative(derivatives[i])
		if err != nil {
			return "", fmt.Errorf("Error pricing option: %v", err)
		}
	}

	derivativesJSON, err := json.Marshal(derivatives)
	if err != nil {
		return "", fmt.Errorf("failed to marshal derivatives to JSON: %w", err)
	}

	return string(derivativesJSON), nil
}

func PriceDerivatives(columnName string){
	var derivatives []DBDerivative
	result := db.DB.Raw("SELECT * FROM derivatives_table;").Scan(&derivatives)
	if result.Error != nil {
		log.Fatalf("Failed to execute raw SQL query: %v", result.Error)
	}
	var err error
	for i := range derivatives {
		derivatives[i].TimeToExpiry = calculateTimeToExpiry(derivatives[i].ExpirationDate)
		derivatives[i].SpotPrice, err = fetchCurrentPrice(strings.ToLower(derivatives[i].AssetName))
		if err != nil {
			fmt.Errorf("Error fetching Spot Price: %v", err)
			return
		}
		derivatives[i].CurrentPrice, err = priceDerivative(derivatives[i])
		if err != nil {
			fmt.Errorf("Error pricing option: %v", err)
			return
		}
		now := time.Now()
		var newPriceTimestamp PriceTimestamp
		newPriceTimestamp.TimestampOfPrice = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, now.Location()).Format("2006-01-02 15:04:05.99999")
		newPriceTimestamp.Price = derivatives[i].CurrentPrice

		sqlStatement := fmt.Sprintf("UPDATE derivatives_table SET %s = array_append(%s, ROW('%s', %f)::PRICETIMESTAMP) WHERE id = '%s';", columnName, columnName, newPriceTimestamp.TimestampOfPrice, newPriceTimestamp.Price, derivatives[i].DerivativeID)
		db.DB.Exec(sqlStatement)
	}
}