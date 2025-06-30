package main

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/petertjmills/escpos-server/daily"
)

func main() {
	// open .env file if it's there and set os env vars
	if err := godotenv.Load(); err != nil {
		fmt.Println("Error loading .env file:", err)
	}

	// dailyWeather := daily.GetWeatherMD()
	// dailyWotd, _ := daily.GetWordOfTheDay()
	dailyNews, _ := daily.GetNews()
	fmt.Println(dailyNews)

}
