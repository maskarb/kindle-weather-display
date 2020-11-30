package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/maskarb/kindle-weather-display/server"
)

var (
	realTimeFields = "precipitation,precipitation_type,temp,feels_like,dewpoint,wind_speed,wind_gust,baro_pressure,visibility,humidity,wind_direction,sunrise,sunset,cloud_cover,cloud_ceiling,cloud_base,surface_shortwave_radiation,moon_phase,weather_code,pm25,pm10,o3,no2,co,so2,epa_aqi,epa_primary_pollutant,epa_health_concern,pollen_tree,pollen_weed,pollen_grass,road_risk_score,road_risk,road_risk_confidence,road_risk_conditions,fire_index,hail_binary"
	hourlyFields   = "precipitation,precipitation_type,precipitation_probability,temp,feels_like,dewpoint,wind_speed,wind_gust,baro_pressure,visibility,humidity,wind_direction,sunrise,sunset,cloud_cover,cloud_ceiling,cloud_base,surface_shortwave_radiation,moon_phase,weather_code,pm25,pm10,o3,no2,co,so2,epa_aqi,epa_primary_pollutant,epa_health_concern,pollen_tree,pollen_weed,pollen_grass,road_risk_score,road_risk,road_risk_confidence,road_risk_conditions,hail_binary"
	dailyFields    = "precipitation,precipitation_accumulation,temp,feels_like,wind_speed,baro_pressure,visibility,humidity,wind_direction,sunrise,sunset,moon_phase,weather_code,dewpoint"
	iconMap        = map[string]string{
		// "clear-day":           "skc",
		// "clear-night":         "skc",
		// "rain":                "ra",
		// "snow":                "sn",
		// "sleet":               "fzra",
		// "wind":                "wind",
		// "fog":                 "fg",
		// "cloudy":              "ovc",
		// "partly-cloudy-day":   "few",
		// "partly-cloudy-night": "few",

		"freezing_rain_heavy": "",
		"freezing_rain":       "",
		"freezing_rain_light": "",
		"freezing_drizzle":    "",
		"ice_pellets_heavy":   "",
		"ice_pellets":         "",
		"ice_pellets_light":   "",
		"snow_heavy":          "",
		"snow":                "",
		"snow_light":          "",
		"flurries":            "",
		"tstorm":              "",
		"rain_heavy":          "",
		"rain":                "",
		"rain_light":          "",
		"drizzle":             "",
		"fog_light":           "",
		"fog":                 "",
		"cloudy":              "",
		"mostly_cloudy":       "",
		"partly_cloudy":       "",
		"mostly_clear":        "",
		"clear":               "",
	}
)

func main() {
	var c *server.Client
	c = server.New(os.Getenv("CLIMACELL_API_KEY"))

	lat := os.Getenv("LATITUDE")
	lon := os.Getenv("LONGITUDE")

	queries := map[string]string{
		"lat": lat,
		"lon": lon,
	}

	queries["fields"] = realTimeFields
	current, err := c.RealTime(queries)
	if err != nil {
		panic(err)
	}

	// queries["fields"] = hourlyFields
	// hours, err := c.HourlyForecast(queries)
	// if err != nil {
	// 	panic(err)
	// }

	queries["fields"] = dailyFields
	daily, err := c.DailyForecast(queries)
	if err != nil {
		panic(err)
	}

	// In our svg file
	// Day One is today
	// Day Two is tomorrow
	// Day three is two days from now
	// day four is three days from now
	// and so forth
	currently := current
	currentDay := daily[0]
	tomorrow := daily[1]
	inTwoDays := daily[2]
	inThreeDays := daily[3]

	tomorrowTime := tomorrow.ObservationTime.Value
	inTwoDaysTime := inTwoDays.ObservationTime.Value
	inThreeDaysTime := inThreeDays.ObservationTime.Value
	location := current.ObservationTime.Value.Location()

	sunrise := currentDay.Sunrise.Value
	sunset := currentDay.Sunset.Value

	updatedTime := time.Now().In(location).Format("Monday Jan 2 15:04")

	svg, err := ioutil.ReadFile("./preprocess.svg")
	if err != nil {
		panic(err)
	}

	svg = bytes.Replace(svg, []byte("TEMP_NOW"), []byte(strconv.FormatFloat(currently.Temperature.Value, 'f', 0, 64)), -1)
	svg = bytes.Replace(svg, []byte("SUNRISE"), []byte(sunrise.Format(time.Kitchen)), -1)
	svg = bytes.Replace(svg, []byte("SUNSET"), []byte(sunset.Format(time.Kitchen)), -1)
	svg = bytes.Replace(svg, []byte("MOON_PHASE"), []byte(currentDay.MoonPhase.Value), -1)
	svg = bytes.Replace(svg, []byte("WIND_SPEED"), []byte(strconv.FormatFloat(currently.WindSpeed.Value, 'f', 0, 64)), -1)
	svg = bytes.Replace(svg, []byte("WIND_DIR"), []byte(strconv.FormatFloat(currently.WindDirection.Value, 'f', 0, 64)), -1)
	svg = bytes.Replace(svg, []byte("HIGH_ONE"), []byte(strconv.FormatFloat(currentDay.Temperature.Max.Max.Value, 'f', 0, 64)), -1)
	svg = bytes.Replace(svg, []byte("HIGH_TWO"), []byte(strconv.FormatFloat(tomorrow.Temperature.Max.Max.Value, 'f', 0, 64)), -1)
	svg = bytes.Replace(svg, []byte("HIGH_THREE"), []byte(strconv.FormatFloat(inTwoDays.Temperature.Max.Max.Value, 'f', 0, 64)), -1)
	svg = bytes.Replace(svg, []byte("HIGH_FOUR"), []byte(strconv.FormatFloat(inThreeDays.Temperature.Max.Max.Value, 'f', 0, 64)), -1)
	svg = bytes.Replace(svg, []byte("LOW_ONE"), []byte(strconv.FormatFloat(currentDay.Temperature.Min.Min.Value, 'f', 0, 64)), -1)
	svg = bytes.Replace(svg, []byte("LOW_TWO"), []byte(strconv.FormatFloat(tomorrow.Temperature.Min.Min.Value, 'f', 0, 64)), -1)
	svg = bytes.Replace(svg, []byte("LOW_THREE"), []byte(strconv.FormatFloat(inTwoDays.Temperature.Min.Min.Value, 'f', 0, 64)), -1)
	svg = bytes.Replace(svg, []byte("LOW_FOUR"), []byte(strconv.FormatFloat(inThreeDays.Temperature.Min.Min.Value, 'f', 0, 64)), -1)
	svg = bytes.Replace(svg, []byte("DAY_TWO"), []byte(tomorrowTime.Weekday().String()), -1)
	svg = bytes.Replace(svg, []byte("DAY_THREE"), []byte(inTwoDaysTime.Weekday().String()), -1)
	svg = bytes.Replace(svg, []byte("DAY_FOUR"), []byte(inThreeDaysTime.Weekday().String()), -1)
	svg = bytes.Replace(svg, []byte("ICON_ONE"), []byte(getIcon(currently.WeatherCode.Value)), -1)
	svg = bytes.Replace(svg, []byte("ICON_TWO"), []byte(getIcon(tomorrow.WeatherCode.Value)), -1)
	svg = bytes.Replace(svg, []byte("ICON_THREE"), []byte(getIcon(inTwoDays.WeatherCode.Value)), -1)
	svg = bytes.Replace(svg, []byte("ICON_FOUR"), []byte(getIcon(inThreeDays.WeatherCode.Value)), -1)
	svg = bytes.Replace(svg, []byte("DATE_STRING"), []byte(updatedTime), -1)

	err = ioutil.WriteFile("output.svg", svg, 0660)
	if err != nil {
		panic(err)
	}

}

func getIcon(i string) string {
	icon, ok := iconMap[i]
	if !ok {
		return "mist"
	}
	return icon
}
