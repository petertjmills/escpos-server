package daily

import (
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"os"
	"strings"
	"time"
)

var WEATHERTEMPLATE = `
#### Weather
| High/feels like | {{.HighTemp}}C/{{.HighFeelsLike}}C |
|---|---|
| Low/feels like | {{.LowTemp}}C/{{.LowFeelsLike}}C |
| Humidity | {{.Humidity}}% |
| Wind | {{.WindSpeed}} mph |
| MSLP (avg/hi/lo) | {{.MSLP}} hPa |
| UV Index | {{.UVIndex}} |
{{range .RainTimes}}
| Rain: {{.}} | {{.Rain}}% |
{{end}}
`

type WeatherRes struct {
	HighTemp      float64            `json:"highTemp"`
	HighFeelsLike float64            `json:"highFeelsLike"`
	LowTemp       float64            `json:"lowTemp"`
	LowFeelsLike  float64            `json:"lowFeelsLike"`
	Humidity      float64            `json:"humidity"`
	WindSpeed     float64            `json:"windSpeed"`
	MSLP          string             `json:"mslp"`
	UVIndex       float64            `json:"uvIndex"`
	RainTimes     map[string]float64 `json:"rainTimes"`
	PrintRain     bool               `json:"printRain"`
}

// Custom time type that can handle the API's time format
type APITime struct {
	time.Time
}

func (t *APITime) UnmarshalJSON(data []byte) error {
	// Remove quotes from the JSON string
	timeStr := strings.Trim(string(data), `"`)

	// Parse the time format from the API: "2025-06-29T20:00Z"
	parsedTime, err := time.Parse("2006-01-02T15:04Z", timeStr)
	if err != nil {
		return err
	}

	t.Time = parsedTime
	return nil
}

type WeatherPoint struct {
	Time                      APITime `json:"time"`
	ScreenTemperature         float64 `json:"screenTemperature"`
	MaxScreenAirTemp          float64 `json:"maxScreenAirTemp"`
	MinScreenAirTemp          float64 `json:"minScreenAirTemp"`
	ScreenDewPointTemperature float64 `json:"screenDewPointTemperature"`
	FeelsLikeTemperature      float64 `json:"feelsLikeTemperature"`
	WindSpeed10M              float64 `json:"windSpeed10m"`
	WindDirectionFrom10M      int     `json:"windDirectionFrom10m"`
	WindGustSpeed10M          float64 `json:"windGustSpeed10m"`
	Max10MWindGust            float64 `json:"max10mWindGust"`
	Visibility                int     `json:"visibility"`
	ScreenRelativeHumidity    float64 `json:"screenRelativeHumidity"`
	Mslp                      int     `json:"mslp"`
	UvIndex                   int     `json:"uvIndex"`
	SignificantWeatherCode    int     `json:"significantWeatherCode"`
	PrecipitationRate         float64 `json:"precipitationRate"`
	TotalPrecipAmount         float64 `json:"totalPrecipAmount"`
	TotalSnowAmount           int     `json:"totalSnowAmount"`
	ProbOfPrecipitation       int     `json:"probOfPrecipitation"`
}

type MetOfficeResp struct {
	Features []struct {
		Properties struct {
			TimeSeries []WeatherPoint `json:"timeSeries"`
		} `json:"properties"`
	} `json:"features"`
}

func getDailyHigh(d MetOfficeResp) (float64, float64) {
	var dailyHigh float64
	var feelsLikeHigh float64
	currentTime := d.Features[0].Properties.TimeSeries[0].Time.Time
	localTime := currentTime.In(time.Local)
	for _, point := range d.Features[0].Properties.TimeSeries {
		if point.Time.Time.In(time.Local).Format("2006-01-02") == localTime.Format("2006-01-02") {
			if point.ScreenTemperature > dailyHigh {
				dailyHigh = point.ScreenTemperature
			}
			if point.FeelsLikeTemperature > feelsLikeHigh {
				feelsLikeHigh = point.FeelsLikeTemperature
			}
		}
	}

	return dailyHigh, feelsLikeHigh
}

func getDailyLow(d MetOfficeResp) (float64, float64) {
	var dailyLow float64
	var feelsLikeLow float64

	dailyLow = math.MaxFloat64
	feelsLikeLow = math.MaxFloat64

	currentTime := d.Features[0].Properties.TimeSeries[0].Time.Time
	localTime := currentTime.In(time.Local)
	for _, point := range d.Features[0].Properties.TimeSeries {
		if point.Time.Time.In(time.Local).Format("2006-01-02") == localTime.Format("2006-01-02") {
			if point.ScreenTemperature < dailyLow {
				dailyLow = point.ScreenTemperature
			}
			if point.FeelsLikeTemperature < feelsLikeLow {
				feelsLikeLow = point.FeelsLikeTemperature
			}
		}
	}

	return dailyLow, feelsLikeLow
}

func getHighHumidity(d MetOfficeResp) float64 {
	var highHumidity float64
	currentTime := d.Features[0].Properties.TimeSeries[0].Time
	localTime := currentTime.In(time.Local)
	for _, point := range d.Features[0].Properties.TimeSeries {
		if point.Time.In(time.Local).Format("2006-01-02") == localTime.Format("2006-01-02") {
			if point.ScreenRelativeHumidity > highHumidity {
				highHumidity = point.ScreenRelativeHumidity
			}
		}
	}

	return highHumidity
}

func getHighWindSpeed(d MetOfficeResp) float64 {
	var highWindSpeed float64
	currentTime := d.Features[0].Properties.TimeSeries[0].Time
	localTime := currentTime.In(time.Local)
	for _, point := range d.Features[0].Properties.TimeSeries {
		if point.Time.In(time.Local).Format("2006-01-02") == localTime.Format("2006-01-02") {
			if point.WindSpeed10M > highWindSpeed {
				highWindSpeed = point.WindSpeed10M
			}
		}
	}

	// from m/s to mph
	return highWindSpeed * 2.237
}

func getAvgMSLP(d MetOfficeResp) (float64, float64, float64) {
	mslpPoints := make([]int, 0)
	var count int
	currentTime := d.Features[0].Properties.TimeSeries[0].Time
	localTime := currentTime.In(time.Local)
	for _, point := range d.Features[0].Properties.TimeSeries {
		if point.Time.In(time.Local).Format("2006-01-02") == localTime.Format("2006-01-02") {
			mslpPoints = append(mslpPoints, point.Mslp)
			count++
		}
	}
	mean := float64(mslpPoints[0])
	for i := 1; i < len(mslpPoints); i++ {
		mean += float64(mslpPoints[i])
	}
	mean /= float64(count)
	low := float64(mslpPoints[0])
	for i := 1; i < len(mslpPoints); i++ {
		if float64(mslpPoints[i]) < low {
			low = float64(mslpPoints[i])
		}
	}
	high := float64(mslpPoints[0])
	for i := 1; i < len(mslpPoints); i++ {
		if float64(mslpPoints[i]) > high {
			high = float64(mslpPoints[i])
		}
	}

	return mean / 100, low / 100, high / 100
}

func getHighUV(d MetOfficeResp) float64 {
	var highUV float64
	currentTime := d.Features[0].Properties.TimeSeries[0].Time
	localTime := currentTime.In(time.Local)
	for _, point := range d.Features[0].Properties.TimeSeries {
		if point.Time.In(time.Local).Format("2006-01-02") == localTime.Format("2006-01-02") {
			if float64(point.UvIndex) > highUV {
				highUV = float64(point.UvIndex)
			}
		}
	}
	return highUV
}

func getRainTimes(d MetOfficeResp) map[string]float64 {
	rainTimes := make(map[string]float64)
	currentTime := d.Features[0].Properties.TimeSeries[0].Time
	localTime := currentTime.In(time.Local)
	for _, point := range d.Features[0].Properties.TimeSeries {
		if point.Time.In(time.Local).Format("2006-01-02") == localTime.Format("2006-01-02") {
			if point.ProbOfPrecipitation > 50 {
				rainTimes[point.Time.In(time.Local).Format("15:04")] = float64(point.ProbOfPrecipitation)
			}
		}
	}
	return rainTimes
}

func GetWeatherMD() string {
	// load api key from env
	key := os.Getenv("WEATHER_API_KEY")
	lat := os.Getenv("WEATHER_LAT")
	lon := os.Getenv("WEATHER_LON")

	// request to  https://data.hub.api.metoffice.gov.uk/sitespecific/v0/point/hourly
	// key should be header
	req, err := http.NewRequest("GET", fmt.Sprintf("https://data.hub.api.metoffice.gov.uk/sitespecific/v0/point/hourly?latitude=%s&longitude=%s", lat, lon), nil)
	req.Header.Add("apikey", key)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Sprintf("Error fetching weather data: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Sprintf("Error fetching weather data: %v", resp.Status)
	}
	defer resp.Body.Close()

	var respData MetOfficeResp
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		return fmt.Sprintf("Error decoding weather data: %v", err)
	}

	var res WeatherRes
	res.HighTemp, res.HighFeelsLike = getDailyHigh(respData)
	res.LowTemp, res.LowFeelsLike = getDailyLow(respData)
	res.Humidity = getHighHumidity(respData)
	res.WindSpeed = float64(int(getHighWindSpeed(respData)))
	mmslp, hmslp, lmslp := getAvgMSLP(respData)
	res.MSLP = fmt.Sprintf("%.f,%.f,%.f", mmslp, hmslp, lmslp)
	res.UVIndex = getHighUV(respData)
	res.RainTimes = getRainTimes(respData)
	if len(res.RainTimes) > 0 {
		res.PrintRain = true
	}

	tmpl, err := template.New("weather").Parse(WEATHERTEMPLATE)
	if err != nil {
		return fmt.Sprintf("Error parsing weather template: %v", err)
	}

	var ret strings.Builder

	if err := tmpl.Execute(&ret, res); err != nil {
		return fmt.Sprintf("Error executing weather template: %v", err)
	}

	return ret.String()
}
