package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/andyhaskell/climacell-go"
)

var (
	// extra          = "pm25,pm10,o3,no2,co,so2,epa_aqi,epa_primary_pollutant,epa_health_concern,pollen_tree,pollen_weed,pollen_grass,road_risk_score,road_risk,road_risk_confidence,road_risk_conditions,fire_index,hail_binary"

	realTimeFields = "precipitation,precipitation_type,temp,feels_like,dewpoint,wind_speed,wind_gust,baro_pressure,visibility,humidity,wind_direction,sunrise,sunset,cloud_cover,cloud_ceiling,cloud_base,surface_shortwave_radiation,moon_phase,weather_code"
	hourlyFields   = "precipitation,precipitation_type,precipitation_probability,temp,feels_like,dewpoint,wind_speed,wind_gust,baro_pressure,visibility,humidity,wind_direction,sunrise,sunset,cloud_cover,cloud_ceiling,cloud_base,surface_shortwave_radiation,moon_phase,weather_code"
	dailyFields    = "precipitation,precipitation_accumulation,temp,feels_like,wind_speed,baro_pressure,visibility,humidity,wind_direction,sunrise,sunset,moon_phase,weather_code,dewpoint"

	iconMap = map[string]string{
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

		"freezing_rain_heavy": "frh",
		"freezing_rain":       "fzra",
		"freezing_rain_light": "frl",
		"freezing_drizzle":    "fd",
		"ice_pellets_heavy":   "iph",
		"ice_pellets":         "ip",
		"ice_pellets_light":   "ipl",
		"snow_heavy":          "sn",
		"snow":                "sn",
		"snow_light":          "sn",
		"flurries":            "sn",
		"tstorm":              "tsra",
		"rain_heavy":          "rh",
		"rain":                "ra",
		"rain_light":          "shra",
		"drizzle":             "d",
		"fog_light":           "sctfg",
		"fog":                 "fg",
		"cloudy":              "ovc",
		"mostly_cloudy":       "bkn",
		"partly_cloudy":       "sct",
		"mostly_clear":        "few",
		"clear":               "skc",
	}
)

func getEnvString(key string, defaultVal string) string {
	valueStr, exists := os.LookupEnv(key)
	if !exists {
		log.Printf("env variable %s not defined", key)
	}
	return valueStr
}

func getEnvAsFloat64(key string, defaultVal float64) float64 {
	valueStr, exists := os.LookupEnv(key)
	if !exists {
		log.Printf("env variable %s not defined", key)
	}
	if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return value
	}
	return defaultVal
}

func main() {
	var c *climacell.Client
	c = climacell.New(getEnvString("CLIMACELL_API_KEY", ""))
	loc := &climacell.LatLon{
		Lat: getEnvAsFloat64("LATITUDE", 0),
		Lon: getEnvAsFloat64("LONGITUDE", 0),
	}

	if err := genFile(c, loc); err != nil {
		log.Printf("output is jacked, probably: %v", err)
	}

	ticker := time.NewTicker(60 * time.Second)

	go func() {
		for {
			select {
			case <-ticker.C:
				if err := genFile(c, loc); err != nil {
					log.Printf("output is jacked, probably: %v", err)
				}
			}
		}
	}()

	fs := http.FileServer(http.Dir("./out"))

	http.Handle("/out/", http.StripPrefix("/out", fs))
	log.Fatal(http.ListenAndServe(":53084", nil))

	fmt.Println("exiting")

}

func genFile(c *climacell.Client, loc *climacell.LatLon) error {
	start := time.Now()

	log.Printf("getting realtime data")
	current, err := c.RealTime(climacell.ForecastArgs{
		Location:   loc,
		UnitSystem: "us",
		Fields:     []string{realTimeFields},
	})
	if err != nil {
		return fmt.Errorf("error getting realTime data: %v", err)
	}

	log.Printf("getting daily forecast data")
	daily, err := c.DailyForecast(climacell.ForecastArgs{
		Location:   loc,
		UnitSystem: "us",
		Fields:     []string{dailyFields},
		Start:      start,
		End:        time.Now().Add(24 * 5 * time.Hour),
	})
	if err != nil {
		return fmt.Errorf("error getting forecast data: %v", err)
	}

	updatedTime := start.Format("Monday Jan 2 15:04")

	svg, err := ioutil.ReadFile("./preprocess.svg")
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}

	today := daily[0]
	tomorrow := daily[1]
	in2days := daily[2]
	in3days := daily[3]

	svg = bytes.Replace(svg, []byte("TEMP_NOW"), []byte(strconv.FormatFloat(*current.Temp.Value, 'f', 0, 64)), -1)
	svg = bytes.Replace(svg, []byte("SUNRISE"), []byte(current.Sunrise.Value.Local().Format(time.Kitchen)), -1)
	svg = bytes.Replace(svg, []byte("SUNSET"), []byte(current.Sunset.Value.Local().Format(time.Kitchen)), -1)
	svg = bytes.Replace(svg, []byte("MOON_PHASE"), []byte(getMoonPhase(*current.MoonPhase.Value)), -1)
	svg = bytes.Replace(svg, []byte("WIND_SPEED"), []byte(strconv.FormatFloat(*current.WindSpeed.Value, 'f', 0, 64)), -1)
	svg = bytes.Replace(svg, []byte("WIND_DIR"), []byte(strconv.FormatFloat(*current.WindDirection.Value, 'f', 0, 64)), -1)
	svg = bytes.Replace(svg, []byte("HIGH_ONE"), []byte(strconv.FormatFloat(*today.Temp.Max().Value.Value, 'f', 0, 64)), -1)
	svg = bytes.Replace(svg, []byte("HIGH_TWO"), []byte(strconv.FormatFloat(*tomorrow.Temp.Max().Value.Value, 'f', 0, 64)), -1)
	svg = bytes.Replace(svg, []byte("HIGH_THREE"), []byte(strconv.FormatFloat(*in2days.Temp.Max().Value.Value, 'f', 0, 64)), -1)
	svg = bytes.Replace(svg, []byte("HIGH_FOUR"), []byte(strconv.FormatFloat(*in3days.Temp.Max().Value.Value, 'f', 0, 64)), -1)
	svg = bytes.Replace(svg, []byte("LOW_ONE"), []byte(strconv.FormatFloat(*today.Temp.Min().Value.Value, 'f', 0, 64)), -1)
	svg = bytes.Replace(svg, []byte("LOW_TWO"), []byte(strconv.FormatFloat(*tomorrow.Temp.Min().Value.Value, 'f', 0, 64)), -1)
	svg = bytes.Replace(svg, []byte("LOW_THREE"), []byte(strconv.FormatFloat(*in2days.Temp.Min().Value.Value, 'f', 0, 64)), -1)
	svg = bytes.Replace(svg, []byte("LOW_FOUR"), []byte(strconv.FormatFloat(*in3days.Temp.Min().Value.Value, 'f', 0, 64)), -1)
	svg = bytes.Replace(svg, []byte("DAY_TWO"), []byte(tomorrow.ObservationTime.Value.Weekday().String()), -1)
	svg = bytes.Replace(svg, []byte("DAY_THREE"), []byte(in2days.ObservationTime.Value.Weekday().String()), -1)
	svg = bytes.Replace(svg, []byte("DAY_FOUR"), []byte(in3days.ObservationTime.Value.Weekday().String()), -1)
	svg = bytes.Replace(svg, []byte("ICON_ONE"), []byte(getIcon(*current.WeatherCode.Value)), -1)
	svg = bytes.Replace(svg, []byte("ICON_TWO"), []byte(getIcon(*tomorrow.WeatherCode.Value)), -1)
	svg = bytes.Replace(svg, []byte("ICON_THREE"), []byte(getIcon(*in2days.WeatherCode.Value)), -1)
	svg = bytes.Replace(svg, []byte("ICON_FOUR"), []byte(getIcon(*in3days.WeatherCode.Value)), -1)
	svg = bytes.Replace(svg, []byte("ICON_MOON"), []byte(*current.MoonPhase.Value), -1)
	svg = bytes.Replace(svg, []byte("DATE_STRING"), []byte(updatedTime), -1)

	f, err := os.OpenFile("out/output.svg", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	if _, err := f.Write(svg); err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}

	log.Printf("writing output to svg")
	if err := f.Close(); err != nil {
		return fmt.Errorf("error closing file: %v", err)
	}

	log.Printf("converting svg to png")
	if err := exec.Command("rsvg-convert", "out/output.svg", "-b", "white", "-f", "png", "-o", "out/output.png").Run(); err != nil {
		return fmt.Errorf("error convert svg to png: %v", err)
	}
	return nil

}

func getIcon(i string) string {
	icon, ok := iconMap[i]
	if !ok {
		return "mist"
	}
	return icon
}

func getMoonPhase(i string) string {
	s := strings.Split(i, "_")
	return strings.Title(s[0])
}
