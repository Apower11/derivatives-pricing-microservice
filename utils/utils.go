package utils

import (
	"net/http"
	"math"
	"time"
	"github.com/Apower11/derivatives-pricing-microservice/db"
	"github.com/Apower11/derivatives-pricing-microservice/types"
	"gorm.io/gorm"
	"errors"
)

func EnableCORS(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, Authorization")
}

func RoundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

func GetStartOfTodayTimestamps() (thisYearTimestamp int64, lastYearTimestamp int64) {
	now := time.Now()
	year := now.Year()

	thisYear := time.Date(year, now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	thisYearTimestamp = thisYear.Unix()

	lastYear := time.Date(year-1, now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	lastYearTimestamp = lastYear.Unix()

	return thisYearTimestamp, lastYearTimestamp
}

func GetUserIDByUsername(username string) (string, error) {
    var user types.User
    result := db.DB.Where("username = ?", username).Select("id").First(&user)

    if result.Error != nil {
        if errors.Is(result.Error, gorm.ErrRecordNotFound) {
            return "", nil
        }
        return "", result.Error
    }
    return user.ID, nil
}