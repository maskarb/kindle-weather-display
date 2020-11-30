package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

type Client struct {
	baseURL string
	apiKey  string
	c       *http.Client
}

type queries map[string]string

func newDefaultHTTPClient() *http.Client { return &http.Client{Timeout: time.Minute} }

func New(apiKey string) *Client { return NewWithClient(apiKey, newDefaultHTTPClient()) }

func NewWithClient(apiKey string, c *http.Client) *Client {
	return &Client{
		baseURL: "https://api.climacell.co/v3/",
		apiKey:  apiKey,
		c:       c,
	}
}

func (c *Client) RealTime(args queries) (*realTime, error) {
	var w realTime
	if err := c.getWeather("weather/realtime", args, &w); err != nil {
		return nil, err
	}
	return &w, nil
}

func (c *Client) HourlyForecast(args queries) (*hourlyWeather, error) {
	var w hourlyWeather
	if err := c.getWeather("weather/forecast/hourly", args, &w); err != nil {
		return nil, err
	}
	return &w, nil
}

func (c *Client) DailyForecast(args queries) ([]daily, error) {
	var w []daily
	if err := c.getWeather("weather/forecast/daily", args, &w); err != nil {
		return nil, err
	}
	return w, nil
}

func (c *Client) getWeather(
	endpt string,
	args queries,
	expectedResponse interface{},
) error {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return errors.WithMessage(err, "parsing base URL")
	}
	u = u.ResolveReference(&url.URL{Path: endpt})

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return errors.WithMessage(err, "making HTTP request")
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("apikey", c.apiKey)
	q := req.URL.Query()
	for k, v := range args {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	res, err := c.c.Do(req)
	if err != nil {
		return errors.WithMessagef(err, "sending weather data request to %s", endpt)
	}
	defer res.Body.Close()

	switch res.StatusCode {
	case 200:
		if err := json.NewDecoder(res.Body).Decode(expectedResponse); err != nil {
			return errors.WithMessage(err, "deserializing weather response data")
		}
		return nil
	case 400, 401, 403, 404, 500:
		var errRes ErrorResponse
		if err := json.NewDecoder(res.Body).Decode(&errRes); err != nil {
			return errors.WithMessage(err, "deserializing weather error response")
		}

		if res.StatusCode == 401 || res.StatusCode == 403 {
			errRes.StatusCode = res.StatusCode
		}
		return &errRes
	default:
		return fmt.Errorf("unexpected HTTP response status code: %d", res.StatusCode)
	}
}

type ErrorResponse struct {
	// StatusCode indicates the HTTP status for this errored API request.
	// For 401 and 403 errors, this is not present in the actual API
	// response's JSON, so this is filled in for us.
	StatusCode int `json:"statusCode"`
	// ErrorCode is the error code for this request. Not present on 401 and
	// 403 errors.
	ErrorCode string `json:"errorCode"`
	// Message is a description of the error that took place.
	Message string `json:"message"`
}

func (err *ErrorResponse) Error() string {
	if err.ErrorCode == "" {
		return fmt.Sprintf("%d API error: %s", err.StatusCode, err.Message)
	}
	return fmt.Sprintf("%d (%s) API error: %s", err.StatusCode, err.ErrorCode, err.Message)
}
