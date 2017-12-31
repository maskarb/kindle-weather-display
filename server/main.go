package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

// DarkSky output of json object from DarkSky API
type DarkSky struct {
	Currently struct {
		ApparentTemperature  float64 `json:"apparentTemperature"`
		CloudCover           float64 `json:"cloudCover"`
		DewPoint             float64 `json:"dewPoint"`
		Humidity             float64 `json:"humidity"`
		Icon                 string  `json:"icon"`
		NearestStormBearing  int64   `json:"nearestStormBearing"`
		NearestStormDistance int64   `json:"nearestStormDistance"`
		Ozone                float64 `json:"ozone"`
		PrecipIntensity      int64   `json:"precipIntensity"`
		PrecipProbability    int64   `json:"precipProbability"`
		Pressure             float64 `json:"pressure"`
		Summary              string  `json:"summary"`
		Temperature          float64 `json:"temperature"`
		Time                 int64   `json:"time"`
		UvIndex              int64   `json:"uvIndex"`
		Visibility           float64 `json:"visibility"`
		WindBearing          uint64  `json:"windBearing"`
		WindGust             float64 `json:"windGust"`
		WindSpeed            float64 `json:"windSpeed"`
	} `json:"currently"`
	Daily struct {
		Data []struct {
			ApparentTemperatureHigh     float64 `json:"apparentTemperatureHigh"`
			ApparentTemperatureHighTime int64   `json:"apparentTemperatureHighTime"`
			ApparentTemperatureLow      float64 `json:"apparentTemperatureLow"`
			ApparentTemperatureLowTime  int64   `json:"apparentTemperatureLowTime"`
			ApparentTemperatureMax      float64 `json:"apparentTemperatureMax"`
			ApparentTemperatureMaxTime  int64   `json:"apparentTemperatureMaxTime"`
			ApparentTemperatureMin      float64 `json:"apparentTemperatureMin"`
			ApparentTemperatureMinTime  int64   `json:"apparentTemperatureMinTime"`
			CloudCover                  float64 `json:"cloudCover"`
			DewPoint                    float64 `json:"dewPoint"`
			Humidity                    float64 `json:"humidity"`
			Icon                        string  `json:"icon"`
			MoonPhase                   float64 `json:"moonPhase"`
			Ozone                       float64 `json:"ozone"`
			PrecipAccumulation          float64 `json:"precipAccumulation"`
			PrecipIntensity             float64 `json:"precipIntensity"`
			PrecipIntensityMax          float64 `json:"precipIntensityMax"`
			PrecipIntensityMaxTime      int64   `json:"precipIntensityMaxTime"`
			PrecipProbability           float64 `json:"precipProbability"`
			PrecipType                  string  `json:"precipType"`
			Pressure                    float64 `json:"pressure"`
			Summary                     string  `json:"summary"`
			SunriseTime                 int64   `json:"sunriseTime"`
			SunsetTime                  int64   `json:"sunsetTime"`
			TemperatureHigh             float64 `json:"temperatureHigh"`
			TemperatureHighTime         int64   `json:"temperatureHighTime"`
			TemperatureLow              float64 `json:"temperatureLow"`
			TemperatureLowTime          int64   `json:"temperatureLowTime"`
			TemperatureMax              float64 `json:"temperatureMax"`
			TemperatureMaxTime          int64   `json:"temperatureMaxTime"`
			TemperatureMin              float64 `json:"temperatureMin"`
			TemperatureMinTime          int64   `json:"temperatureMinTime"`
			Time                        int64   `json:"time"`
			UvIndex                     int64   `json:"uvIndex"`
			UvIndexTime                 int64   `json:"uvIndexTime"`
			Visibility                  float64 `json:"visibility"`
			WindBearing                 uint64  `json:"windBearing"`
			WindGust                    float64 `json:"windGust"`
			WindGustTime                int64   `json:"windGustTime"`
			WindSpeed                   float64 `json:"windSpeed"`
		} `json:"data"`
		Icon    string `json:"icon"`
		Summary string `json:"summary"`
	} `json:"daily"`
	Flags struct {
		IsdStations []string `json:"isd-stations"`
		Sources     []string `json:"sources"`
		Units       string   `json:"units"`
	} `json:"flags"`
	Hourly struct {
		Data []struct {
			ApparentTemperature float64 `json:"apparentTemperature"`
			CloudCover          float64 `json:"cloudCover"`
			DewPoint            float64 `json:"dewPoint"`
			Humidity            float64 `json:"humidity"`
			Icon                string  `json:"icon"`
			Ozone               float64 `json:"ozone"`
			PrecipAccumulation  float64 `json:"precipAccumulation"`
			PrecipIntensity     float64 `json:"precipIntensity"`
			PrecipProbability   float64 `json:"precipProbability"`
			PrecipType          string  `json:"precipType"`
			Pressure            float64 `json:"pressure"`
			Summary             string  `json:"summary"`
			Temperature         float64 `json:"temperature"`
			Time                int64   `json:"time"`
			UvIndex             int64   `json:"uvIndex"`
			Visibility          int64   `json:"visibility"`
			WindBearing         uint64  `json:"windBearing"`
			WindGust            float64 `json:"windGust"`
			WindSpeed           float64 `json:"windSpeed"`
		} `json:"data"`
		Icon    string `json:"icon"`
		Summary string `json:"summary"`
	} `json:"hourly"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Minutely  struct {
		Data []struct {
			PrecipIntensity   float64 `json:"precipIntensity"`
			PrecipProbability float64 `json:"precipProbability"`
			Time              int64   `json:"time"`
		} `json:"data"`
		Icon    string `json:"icon"`
		Summary string `json:"summary"`
	} `json:"minutely"`
	Offset   int64  `json:"offset"`
	Timezone string `json:"timezone"`
}

func getWeather() io.ReadCloser {
	c := http.Client{
		Timeout: 10 * time.Second,
	}

	s := fmt.Sprintf("https://api.darksky.net/forecast/%s/%s", os.Getenv("DARKSKY_API_KEY"), os.Getenv("GPS_COORDINATES"))
	r, err := c.Get(s)
	if err != nil {
		panic(err)
	}
	return r.Body
}

func main() {
	response, err := ioutil.ReadAll(getWeather())
	if err != nil {
		panic(err)
	}

	weather := DarkSky{}

	json.Unmarshal(response, &weather)
	if err != nil {
		panic(err)
	}

	// In our svg file
	// Day One is today
	// Day Two is tomorrow
	// Day three is two days from now
	// day four is three days from now
	// and so forth
	currently := weather.Currently
	currentDay := weather.Daily.Data[0]
	tomorrow := weather.Daily.Data[1]
	inTwoDays := weather.Daily.Data[2]
	inThreeDays := weather.Daily.Data[3]

	tomorrowTime := time.Unix(tomorrow.Time, 0)
	inTwoDaysTime := time.Unix(inTwoDays.Time, 0)
	inThreeDaysTime := time.Unix(inThreeDays.Time, 0)
	location, err := time.LoadLocation(weather.Timezone)
	if err != nil {
		panic(err)
	}
	sunrise := time.Unix(currentDay.SunriseTime, 0).In(location)
	sunset := time.Unix(currentDay.SunsetTime, 0).In(location)

	updatedTime := time.Now().In(location).Format("Monday Jan 2 15:04")

	svg, err := ioutil.ReadFile("./preprocess.svg")
	if err != nil {
		panic(err)
	}

	svg = bytes.Replace(svg, []byte("TEMP_NOW"), []byte(strconv.FormatFloat(currently.Temperature, 'f', 0, 32)), -1)
	svg = bytes.Replace(svg, []byte("SUNRISE"), []byte(sunrise.Format(time.Kitchen)), -1)
	svg = bytes.Replace(svg, []byte("SUNSET"), []byte(sunset.Format(time.Kitchen)), -1)
	svg = bytes.Replace(svg, []byte("MOON_PHASE"), []byte(moonPhase(currentDay.MoonPhase)), -1)
	svg = bytes.Replace(svg, []byte("WIND_SPEED"), []byte(strconv.FormatFloat(currently.WindSpeed, 'f', 0, 32)), -1)
	svg = bytes.Replace(svg, []byte("WIND_DIR"), []byte(strconv.FormatUint(currently.WindBearing, 10)), -1)
	svg = bytes.Replace(svg, []byte("HIGH_ONE"), []byte(strconv.FormatFloat(currentDay.TemperatureHigh, 'f', 0, 32)), -1)
	svg = bytes.Replace(svg, []byte("HIGH_TWO"), []byte(strconv.FormatFloat(tomorrow.TemperatureHigh, 'f', 0, 32)), -1)
	svg = bytes.Replace(svg, []byte("HIGH_THREE"), []byte(strconv.FormatFloat(inTwoDays.TemperatureHigh, 'f', 0, 32)), -1)
	svg = bytes.Replace(svg, []byte("HIGH_FOUR"), []byte(strconv.FormatFloat(inThreeDays.TemperatureHigh, 'f', 0, 32)), -1)
	svg = bytes.Replace(svg, []byte("LOW_ONE"), []byte(strconv.FormatFloat(currentDay.TemperatureLow, 'f', 0, 32)), -1)
	svg = bytes.Replace(svg, []byte("LOW_TWO"), []byte(strconv.FormatFloat(tomorrow.TemperatureLow, 'f', 0, 32)), -1)
	svg = bytes.Replace(svg, []byte("LOW_THREE"), []byte(strconv.FormatFloat(inTwoDays.TemperatureLow, 'f', 0, 32)), -1)
	svg = bytes.Replace(svg, []byte("LOW_FOUR"), []byte(strconv.FormatFloat(inThreeDays.TemperatureLow, 'f', 0, 32)), -1)
	svg = bytes.Replace(svg, []byte("DAY_TWO"), []byte(tomorrowTime.Weekday().String()), -1)
	svg = bytes.Replace(svg, []byte("DAY_THREE"), []byte(inTwoDaysTime.Weekday().String()), -1)
	svg = bytes.Replace(svg, []byte("DAY_FOUR"), []byte(inThreeDaysTime.Weekday().String()), -1)
	svg = bytes.Replace(svg, []byte("ICON_ONE"), []byte(iconMap(currently.Icon)), -1)
	svg = bytes.Replace(svg, []byte("ICON_TWO"), []byte(iconMap(tomorrow.Icon)), -1)
	svg = bytes.Replace(svg, []byte("ICON_THREE"), []byte(iconMap(inTwoDays.Icon)), -1)
	svg = bytes.Replace(svg, []byte("ICON_FOUR"), []byte(iconMap(inThreeDays.Icon)), -1)
	svg = bytes.Replace(svg, []byte("DATE_STRING"), []byte(updatedTime), -1)

	err = ioutil.WriteFile("output.svg", svg, 0660)
	if err != nil {
		panic(err)
	}

}

func iconMap(i string) string {
	iconMap := map[string]string{
		"clear-day":           "skc",
		"clear-night":         "skc",
		"rain":                "ra",
		"snow":                "sn",
		"sleet":               "fzra",
		"wind":                "wind",
		"fog":                 "fg",
		"cloudy":              "ovc",
		"partly-cloudy-day":   "few",
		"partly-cloudy-night": "few",
	}

	for item := range iconMap {
		if item == i {
			return iconMap[i]
		}
	}
	return "mist"
}

func moonPhase(i float64) string {
	if i == 0 {
		return "New"
	} else if i > 0 && i < 0.25 {
		return "Waxing Crescent"
	} else if i == 0.25 {
		return "First Quarter"
	} else if i > 0.25 && i < 0.5 {
		return "Waxing Gibbous"
	} else if i == 0.5 {
		return "Full"
	} else if i > 0.5 && i < 0.75 {
		return "Waning Gibbous"
	} else if i == 0.75 {
		return "Last Quarter"
	} else if i > 0.75 {
		return "Waning Crescent"
	}
	return "Fail"
}
