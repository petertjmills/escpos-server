package main

import (
	"fmt"
	"time"

	"github.com/joho/godotenv"
	"github.com/petertjmills/escpos-server/daily"
)

func main() {
	// open .env file if it's there and set os env vars
	if err := godotenv.Load(); err != nil {
		fmt.Println("Error loading .env file:", err)
	}
	// print current day like monday 25th July 2025
	fmt.Println("### ", time.Now().Format("Monday, 2 January 2006"))

	dailyWotd, _ := daily.GetWordOfTheDay()
	fmt.Println(dailyWotd)
	dailyWeather := daily.GetWeatherMD()
	fmt.Println(dailyWeather)
	dailyNews, _ := daily.GetNews()
	fmt.Println(dailyNews)
	pollen, _ := daily.GetPollenCount()
	fmt.Println(pollen)
	hn, _ := daily.GetHackerNewsFront()
	fmt.Println(hn)
}
