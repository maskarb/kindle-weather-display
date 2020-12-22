package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/andyhaskell/climacell-go"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
)

var (
	location    *time.Location
	defaultCron = "*/5 * * * *"
	// extra          = "pm25,pm10,o3,no2,co,so2,epa_aqi,epa_primary_pollutant,epa_health_concern,pollen_tree,pollen_weed,pollen_grass,road_risk_score,road_risk,road_risk_confidence,road_risk_conditions,fire_index,hail_binary"

	realTimeFields = "precipitation,precipitation_type,temp,feels_like,dewpoint,wind_speed,wind_gust,baro_pressure,visibility,humidity,wind_direction,sunrise,sunset,cloud_cover,cloud_ceiling,cloud_base,surface_shortwave_radiation,moon_phase,weather_code"
	hourlyFields   = "precipitation,precipitation_type,precipitation_probability,temp,feels_like,dewpoint,wind_speed,wind_gust,baro_pressure,visibility,humidity,wind_direction,sunrise,sunset,cloud_cover,cloud_ceiling,cloud_base,surface_shortwave_radiation,moon_phase,weather_code"
	dailyFields    = "precipitation,precipitation_accumulation,temp,feels_like,wind_speed,baro_pressure,visibility,humidity,wind_direction,sunrise,sunset,moon_phase,weather_code,dewpoint"

	dayNight = []string{"clear", "mostly_clear", "partly_cloudy"}
	iconMap  = map[string]string{
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
		"rain":                "rain",
		"rain_light":          "shra",
		"drizzle":             "d",
		"fog_light":           "sctfg",
		"fog":                 "fg",
		"cloudy":              "cloudy",
		"mostly_cloudy":       "bkn",
		"partly_cloudy":       "sct",
		"mostly_clear":        "few",
		"clear":               "clear_day",
	}
)

func getEnvString(key string, defaultVal string) string {
	valueStr, exists := os.LookupEnv(key)
	if !exists {
		logrus.Infof("env variable %s not defined. Using default: %s", key, defaultVal)
		return defaultVal
	}
	return valueStr
}

func getEnvAsFloat64(key string, defaultVal float64) float64 {
	valueStr, exists := os.LookupEnv(key)
	if !exists {
		logrus.Infof("env variable %s not defined. Using default: %d", key, defaultVal)
	}
	if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return value
	}
	return defaultVal
}

func getDayOrNight(current, rise, set time.Time) string {
	if rise.Before(current) && current.Before(set) {
		return "day"
	}
	return "night"
}

func getWeatherIcon(i, daytime string) string {
	for _, v := range dayNight {
		if i == v {
			return i + "_" + daytime
		}
	}
	return i
}

func getMoonPhase(i string) string {
	s := strings.Replace(i, "_", " ", -1)
	return strings.Title(s)
}

func validateCronSpec(spec string) cron.Schedule {
	s, err := cron.ParseStandard(spec)
	if err != nil {
		logrus.Infof("using default schedule `%s`. Cron spec schedule `%s` not valid: %v", defaultCron, spec, err)
		s, err = cron.ParseStandard(defaultCron)
		if err != nil {
			logrus.Fatalf("failed to parse default schedule: %v", err)
		}
	}
	return s
}

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{})
	tz := getEnvString("TIMEZONE", "UTC")

	var err error
	location, err = time.LoadLocation(tz)
	if err != nil {
		logrus.Infof("defaulting to UTC: timezone %s not found: %v", tz, err)
		location = time.UTC
	}
}

func main() {
	strSpec := getEnvString("CRON_SCHEDULE", defaultCron)
	schedule := validateCronSpec(strSpec)

	filegen := &FileGenerator{
		c: climacell.New(getEnvString("CLIMACELL_API_KEY", "")),
		loc: &climacell.LatLon{
			Lat: getEnvAsFloat64("LATITUDE", 35.780361),
			Lon: getEnvAsFloat64("LONGITUDE", -78.639111),
		},
		sched: schedule,
	}

	if err := filegen.genFile(); err != nil {
		logrus.Info("output is jacked, probably: %v", err)
	}

	cron := cron.New()
	cron.Schedule(filegen.sched, filegen)
	logrus.Infof("starting cronjob on schedule: %s", strSpec)
	cron.Start()

	fs := http.FileServer(http.Dir("./out"))
	http.Handle("/out/", http.StripPrefix("/out", fs))
	logrus.Fatal(http.ListenAndServe(":53084", nil))
	logrus.Info("exiting")
}

type FileGenerator struct {
	c     *climacell.Client
	loc   *climacell.LatLon
	sched cron.Schedule
}

func (f *FileGenerator) Run() {
	if err := f.genFile(); err != nil {
		logrus.Errorf("failed to generate file: %v", err)
	}
}

func (f *FileGenerator) genFile() error {
	start := time.Now()

	logrus.Info("getting realtime data")
	current, err := f.c.RealTime(climacell.ForecastArgs{
		Location:   f.loc,
		UnitSystem: "us",
		Fields:     []string{realTimeFields},
	})
	if err != nil {
		return fmt.Errorf("error getting realTime data: %v", err)
	}

	dayOrNight := getDayOrNight(start, *current.Sunrise.Value, *current.Sunset.Value)

	logrus.Info("getting daily forecast data")
	daily, err := f.c.DailyForecast(climacell.ForecastArgs{
		Location:   f.loc,
		UnitSystem: "us",
		Fields:     []string{dailyFields},
		Start:      start,
		End:        time.Now().Add(24 * 5 * time.Hour),
	})
	if err != nil {
		return fmt.Errorf("error getting forecast data: %v", err)
	}

	updatedTime := start.In(location).Format("Monday Jan 2, 15:04 MST")

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
	svg = bytes.Replace(svg, []byte("ICON_ONE"), []byte(getWeatherIcon(*current.WeatherCode.Value, dayOrNight)), -1)
	svg = bytes.Replace(svg, []byte("ICON_TWO"), []byte(*tomorrow.WeatherCode.Value), -1)
	svg = bytes.Replace(svg, []byte("ICON_THREE"), []byte(*in2days.WeatherCode.Value), -1)
	svg = bytes.Replace(svg, []byte("ICON_FOUR"), []byte(*in3days.WeatherCode.Value), -1)
	svg = bytes.Replace(svg, []byte("ICON_MOON"), []byte(*current.MoonPhase.Value), -1)
	svg = bytes.Replace(svg, []byte("LATITUDE"), []byte(strconv.FormatFloat(f.loc.Lat, 'f', 3, 64)), -1)
	svg = bytes.Replace(svg, []byte("LONGITUDE"), []byte(strconv.FormatFloat(f.loc.Lon, 'f', 3, 64)), -1)
	svg = bytes.Replace(svg, []byte("DATE_STRING"), []byte(updatedTime), -1)

	if _, err := os.Stat("out"); os.IsNotExist(err) {
		logrus.Info("creating `out` folder")
		if err := os.Mkdir("out", 0777); err != nil {
			return fmt.Errorf("cannot create `out` folder: %v", err)
		}
	}

	output, err := os.OpenFile("out/output.svg", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	if _, err := output.Write(svg); err != nil {
		return fmt.Errorf("error writing file: %v", err)
	}

	logrus.Info("writing output to svg")
	if err := output.Close(); err != nil {
		return fmt.Errorf("error closing file: %v", err)
	}

	logrus.Info("converting svg to png")
	if err := exec.Command("rsvg-convert", "out/output.svg", "-b", "white", "-f", "png", "-o", "out/output.png").Run(); err != nil {
		return fmt.Errorf("error convert svg to png: %v", err)
	}
	logrus.Info("created .png output")
	return nil

}
