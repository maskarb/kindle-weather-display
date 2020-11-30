package server

import (
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
	"time"
)

type minimum struct {
	ObservationTime stringV `json:"observation_time"`
	Min             measure `json:"min"`
}

type maximum struct {
	ObservationTime stringV `json:"observation_time"`
	Max             measure `json:"max"`
}

type minMax struct {
	Min minimum
	Max maximum
}

func (mm *minMax) UnmarshalJSON(b []byte) error {
	var records []json.RawMessage
	if err := json.Unmarshal(b, &records); err != nil {
		return err
	}

	if len(records) < 2 {
		return errors.New("short JSON array")
	}

	if err := json.Unmarshal(records[0], &mm.Min); err != nil {
		return err
	}

	if err := json.Unmarshal(records[1], &mm.Max); err != nil {
		return err
	}

	return nil
}

type maxList struct {
	Max maximum
}

func (ml *maxList) UnmarshalJSON(b []byte) error {
	var records []json.RawMessage
	if err := json.Unmarshal(b, &records); err != nil {
		return err
	}

	if err := json.Unmarshal(records[0], &ml.Max); err != nil {
		return err
	}

	return nil
}

// measure is a value (float64) + unit (string)
type measure struct {
	Value float64 `json:"value"`
	Units string  `json:"unit"`
}

// stringMeasure is a value (string) + unit (string)
type stringMeasure struct {
	Value string `json:"value"`
	Units string `json:"unit"`
}

// floatV is a value (float64)
type floatV struct {
	Value float64 `json:"value"`
}

// stringV is a value (string)
type stringV struct {
	Value string `json:"value"`
}

type timeValue struct {
	Value time.Time `json:"value"`
}

type core struct {
	Precipitation             measure   `json:"precipitation,omitempty"`
	PrecipitationType         stringV   `json:"precipitation_type,omitempty"`
	PrecipitationProbability  measure   `json:"precipitation_probability,omitempty"`
	PrecipitationAccumulation measure   `json:"precipitation_accumulation,omitempty"`
	Temperature               measure   `json:"temp,omitempty"`
	FeelsLike                 measure   `json:"feels_like,omitempty"`
	Dewpoint                  measure   `json:"dewpoint,omitempty"`
	WindSpeed                 measure   `json:"wind_speed,omitempty"`
	WindGust                  measure   `json:"wind_gust,omitempty"`
	BaroPressure              measure   `json:"baro_pressure,omitempty"`
	Visibility                measure   `json:"visibility,omitempty"`
	Humidity                  measure   `json:"humidity,omitempty"`
	WindDirection             measure   `json:"wind_direction,omitempty"`
	CloudCover                measure   `json:"cloud_cover,omitempty"`
	CloudCeiling              measure   `json:"cloud_ceiling,omitempty"`
	CloudBase                 measure   `json:"cloud_base,omitempty"`
	SurfaceShortwaveRadiation measure   `json:"surface_shortwave_radiation,omitempty"`
	Sunrise                   timeValue `json:"sunrise,omitempty"`
	Sunset                    timeValue `json:"sunset,omitempty"`
	MoonPhase                 stringV   `json:"moon_phase,omitempty"`
	WeatherCode               stringV   `json:"weather_code,omitempty"`
}

type airQuality struct {
	PM25                measure `json:"pm25,omitempty"`
	PM10                measure `json:"pm10,omitempty"`
	O3                  measure `json:"o3,omitempty"`
	NO2                 measure `json:"no2,omitempty"`
	CO                  measure `json:"co,omitempty"`
	SO2                 measure `json:"so2,omitempty"`
	EPQAQI              floatV  `json:"epa_aqi,omitempty"`
	EPAPrimaryPollutant stringV `json:"epa_primary_pollutant,omitempty"`
	EPAHealthConcern    stringV `json:"epa_health_concern,omitempty"`
}

type pollen struct {
	PollenTree  measure `json:"pollen_tree,omitempty"`
	PollenWeed  measure `json:"pollen_weed,omitempty"`
	PollenGrass measure `json:"pollen_grass,omitempty"`
}

type road struct {
	RoadRiskScore      stringMeasure `json:"road_risk_score,omitempty"`
	RoadRisk           stringV       `json:"road_risk,omitempty"`
	RoadRiskConfidence measure       `json:"road_risk_confidence,omitempty"`
	RoadRiskConditions stringMeasure `json:"road_risk_conditions,omitempty"`
}

type fire struct {
	FireIndex floatV `json:"fire_index,omitempty"`
}

type insurance struct {
	HailBinary measure `json:"hail_binary,omitempty"`
}

type realTime struct {
	*latLon
	ObservationTime timeValue `json:"observation_time"`

	*core
	*airQuality
	*pollen
	*road
	*fire
	*insurance
}

type hourlyWeather struct {
	*latLon
	ObservationTime timeValue `json:"observation_time"`

	*core
	*airQuality
	*pollen
	*road
	*fire
	*insurance
}

type daily struct {
	*latLon
	ObservationTime timeValue `json:"observation_time"`

	Precipitation             maxList   `json:"precipitation"`
	PrecipitationAccumulation measure   `json:"precipitation_accumulation"`
	Temperature               minMax    `json:"temp"`
	FeelsLike                 minMax    `json:"feels_like"`
	WindSpeed                 minMax    `json:"wind_speed"`
	BaroPressure              minMax    `json:"baro_pressure"`
	Visibility                minMax    `json:"visibility"`
	Humidity                  minMax    `json:"humidity"`
	WindDirection             minMax    `json:"wind_direction"`
	Sunrise                   timeValue `json:"sunrise"`
	Sunset                    timeValue `json:"sunset"`
	MoonPhase                 stringV   `json:"moon_phase"`
	WeatherCode               stringV   `json:"weather_code"`
	Dewpoint                  minMax    `json:"dewpoint"`
}

type latLon struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

func (l latLon) locationQueryParams() url.Values {
	return url.Values{
		"lat": []string{strconv.FormatFloat(l.Lat, 'f', -1, 64)},
		"lon": []string{strconv.FormatFloat(l.Lon, 'f', -1, 64)},
	}
}
