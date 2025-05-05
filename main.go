package main

import (
	"net/http"
	"github.com/gorilla/mux"
	"github.com/Apower11/derivatives-pricing-microservice/services/derivatives"
	"github.com/Apower11/derivatives-pricing-microservice/services/auth"
	"github.com/Apower11/derivatives-pricing-microservice/services/portfolio"
	"github.com/Apower11/derivatives-pricing-microservice/services/transactions"
	"github.com/Apower11/derivatives-pricing-microservice/services/assets"
	"github.com/Apower11/derivatives-pricing-microservice/services/posts"
	"github.com/Apower11/derivatives-pricing-microservice/services/users"
	"github.com/Apower11/derivatives-pricing-microservice/services/followers"
	"github.com/Apower11/derivatives-pricing-microservice/services/chats"
	"github.com/Apower11/derivatives-pricing-microservice/services/messages"
	"github.com/Apower11/derivatives-pricing-microservice/db"
	"github.com/joho/godotenv"
)

func main() {
	db.InitDB()
    godotenv.Load()
	go derivatives.RunEveryFifthMinuteFromBeginningOfDay(func() {
		derivatives.PriceDerivatives("five_minute_timestamps");
	})
	go derivatives.RunEveryHalfHour(func() {
		derivatives.PriceDerivatives("thirty_minute_timestamps");
	})
	go derivatives.RunAtMidnight(func(){
		derivatives.PriceDerivatives("one_day_timestamps");
	})
	go derivatives.RunAtStartOfWeek(func(){
		derivatives.PriceDerivatives("one_week_timestamps");
	})
    r := mux.NewRouter()
	assets.RegisterRoutes(r)
	derivatives.RegisterRoutes(r)
	auth.RegisterRoutes(r)
	posts.RegisterRoutes(r)
	portfolio.RegisterRoutes(r)
	transactions.RegisterRoutes(r)
	users.RegisterRoutes(r)
	followers.RegisterRoutes(r)
	chats.RegisterRoutes(r)
	messages.RegisterRoutes(r)
	http.ListenAndServe(":8081", r)
}