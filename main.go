package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// These are the Open-Meteo API endpoints used by the server.
const (
	geocodingURL = "https://geocoding-api.open-meteo.com/v1/search"
	forecastURL  = "https://api.open-meteo.com/v1/forecast"
	forecastDays = 7
)

// WeatherInput is the argument shape accepted by the MCP weather tool.
// The JSON tag names the field sent by MCP clients, and the jsonschema tag
// provides a short description when the SDK builds the tool's input schema.
type WeatherInput struct {
	Location                string `json:"location" jsonschema:"city or place name"`
	IncludeSevenDayForecast bool   `json:"include_seven_day_forecast,omitempty" jsonschema:"Set true only after the user asks for the seven-day forecast"`
}

// WeatherOutput is the data returned by the MCP tool.
// Its JSON and jsonschema tags control the output field names and descriptions.
type WeatherOutput struct {
	Location           string          `json:"location" jsonschema:"Resolved place name"`
	Country            string          `json:"country,omitempty" jsonschema:"Country, when available"`
	TemperatureCelsius float64         `json:"temperature_celsius" jsonschema:"Current air temperature in Celsius"`
	FeelsLikeCelsius   float64         `json:"feels_like_celsius" jsonschema:"Current apparent temperature in Celsius"`
	RelativeHumidity   int             `json:"relative_humidity_percent" jsonschema:"Current relative humidity percentage"`
	WindSpeedKmh       float64         `json:"wind_speed_kmh" jsonschema:"Current wind speed in kilometers per hour"`
	WeatherCode        int             `json:"weather_code" jsonschema:"WMO weather interpretation code"`
	TodayForecast      DailyForecast   `json:"today_forecast" jsonschema:"Forecast for today"`
	Forecast           []DailyForecast `json:"forecast_7_days,omitempty" jsonschema:"Daily weather forecast for the next seven days, when requested"`
	FollowUpQuestion   string          `json:"follow_up_question,omitempty" jsonschema:"Ask the user this question when present"`
}

type DailyForecast struct {
	Date                        string  `json:"date" jsonschema:"Forecast date"`
	MinimumTemperatureCelsius   float64 `json:"minimum_temperature_celsius" jsonschema:"Minimum temperature in Celsius"`
	MaximumTemperatureCelsius   float64 `json:"maximum_temperature_celsius" jsonschema:"Maximum temperature in Celsius"`
	PrecipitationProbabilityPct int     `json:"precipitation_probability_percent" jsonschema:"Maximum daily precipitation probability"`
	WeatherCode                 int     `json:"weather_code" jsonschema:"WMO weather interpretation code"`
}

// geocodingResponse contains only the fields this application needs from
// Open-Meteo's geocoding JSON response. The field tags map JSON names to Go
// fields; the anonymous struct describes each item in the "results" array.
type geocodingResponse struct {
	Results []struct {
		Name      string  `json:"name"`
		Country   string  `json:"country"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"results"`
}

// forecastResponse contains the current and daily forecast fields used by the
// application. Unused fields in Open-Meteo's JSON are safely ignored.
type forecastResponse struct {
	Current struct {
		Temperature      float64 `json:"temperature_2m"`
		FeelsLike        float64 `json:"apparent_temperature"`
		RelativeHumidity int     `json:"relative_humidity_2m"`
		WindSpeed        float64 `json:"wind_speed_10m"`
		WeatherCode      int     `json:"weather_code"`
	} `json:"current"`
	Daily forecastDailyResponse `json:"daily"`
}

type forecastDailyResponse struct {
	Time                     []string  `json:"time"`
	TemperatureMin           []float64 `json:"temperature_2m_min"`
	TemperatureMax           []float64 `json:"temperature_2m_max"`
	PrecipitationProbability []int     `json:"precipitation_probability_max"`
	WeatherCode              []int     `json:"weather_code"`
}

// httpClient is shared by requests to the Open-Meteo APIs. The timeout prevents
// a request from hanging indefinitely if the remote service is unresponsive.
var httpClient = &http.Client{Timeout: 10 * time.Second}

// getCurrentWeather is the handler registered as the MCP tool.
// It resolves the requested place and returns today's forecast. It only
// includes the seven-day forecast when the user has requested it.
//
// ctx is provided by the MCP SDK and lets the client cancel the operation.
// The request parameter is unused by this handler, so it is named "_" .
// The three return values follow the SDK's tool-handler convention: optional
// MCP result metadata, the typed tool output, and an error.
func getCurrentWeather(ctx context.Context, _ *mcp.CallToolRequest, input WeatherInput) (
	*mcp.CallToolResult,
	WeatherOutput,
	error,
) {
	startedAt := time.Now()
	// Remove whitespace from either end so a location such as " Lisbon "
	// is accepted like "Lisbon".
	location := strings.TrimSpace(input.Location)
	log.Printf("get_current_weather request received for %q", location)
	if location == "" {
		log.Printf("get_current_weather rejected an empty location")
		return nil, WeatherOutput{}, fmt.Errorf("location must not be empty")
	}

	// Build the geocoding query as URL parameters. Values are encoded safely
	// (including spaces and accented characters) by url.Values.Encode.
	geocodingQuery := url.Values{
		"name":     {location},
		"count":    {"1"},
		"language": {"en"},
		"format":   {"json"},
	}
	// fetchJSON makes the HTTP request and decodes the response into the
	// geocodingResponse type. The generic type argument tells it the JSON shape.
	geocoding, err := fetchJSON[geocodingResponse](ctx, geocodingURL+"?"+geocodingQuery.Encode())
	if err != nil {
		// Add context to the underlying error while preserving it for callers.
		log.Printf("get_current_weather geocoding failed for %q: %v", location, err)
		return nil, WeatherOutput{}, fmt.Errorf("find location: %w", err)
	}
	// A valid API response can still contain no matches for the requested name.
	if len(geocoding.Results) == 0 {
		log.Printf("get_current_weather found no matching location for %q", location)
		return nil, WeatherOutput{}, fmt.Errorf("no matching location found for %q", location)
	}
	// count=1 was requested, so use the first (and normally only) match.
	place := geocoding.Results[0]

	daysToRequest := 1
	if input.IncludeSevenDayForecast {
		daysToRequest = forecastDays
	}

	// Request current conditions and only the daily forecast data needed for
	// this interaction.
	forecastQuery := url.Values{
		"latitude":         {fmt.Sprintf("%g", place.Latitude)},
		"longitude":        {fmt.Sprintf("%g", place.Longitude)},
		"current":          {"temperature_2m,apparent_temperature,relative_humidity_2m,wind_speed_10m,weather_code"},
		"daily":            {"weather_code,temperature_2m_min,temperature_2m_max,precipitation_probability_max"},
		"forecast_days":    {fmt.Sprintf("%d", daysToRequest)},
		"temperature_unit": {"celsius"},
		"wind_speed_unit":  {"kmh"},
		"timezone":         {"auto"},
	}
	forecast, err := fetchJSON[forecastResponse](ctx, forecastURL+"?"+forecastQuery.Encode())
	if err != nil {
		log.Printf("get_current_weather forecast failed for %q: %v", location, err)
		return nil, WeatherOutput{}, fmt.Errorf("get weather forecast: %w", err)
	}
	dailyForecast, err := mapDailyForecast(forecast.Daily, daysToRequest)
	if err != nil {
		log.Printf("get_current_weather forecast data invalid for %q: %v", location, err)
		return nil, WeatherOutput{}, err
	}

	// Convert the API response types into the public MCP output. The first
	// return value is nil because no extra MCP result metadata is needed.
	output := WeatherOutput{
		Location:           place.Name,
		Country:            place.Country,
		TemperatureCelsius: forecast.Current.Temperature,
		FeelsLikeCelsius:   forecast.Current.FeelsLike,
		RelativeHumidity:   forecast.Current.RelativeHumidity,
		WindSpeedKmh:       forecast.Current.WindSpeed,
		WeatherCode:        forecast.Current.WeatherCode,
		TodayForecast:      dailyForecast[0],
	}
	if input.IncludeSevenDayForecast {
		output.Forecast = dailyForecast
	} else {
		output.FollowUpQuestion = "Would you like the 7-day forecast for " + place.Name + "?"
	}
	log.Printf("get_current_weather completed for %q in %s", place.Name, time.Since(startedAt).Round(time.Millisecond))
	return nil, output, nil
}

func mapDailyForecast(daily forecastDailyResponse, expectedDays int) ([]DailyForecast, error) {
	count := len(daily.Time)
	if count != expectedDays ||
		len(daily.TemperatureMin) != count ||
		len(daily.TemperatureMax) != count ||
		len(daily.PrecipitationProbability) != count ||
		len(daily.WeatherCode) != count {
		return nil, fmt.Errorf("forecast response contains inconsistent daily data")
	}

	forecast := make([]DailyForecast, count)
	for i, date := range daily.Time {
		forecast[i] = DailyForecast{
			Date:                        date,
			MinimumTemperatureCelsius:   daily.TemperatureMin[i],
			MaximumTemperatureCelsius:   daily.TemperatureMax[i],
			PrecipitationProbabilityPct: daily.PrecipitationProbability[i],
			WeatherCode:                 daily.WeatherCode[i],
		}
	}
	return forecast, nil
}

// fetchJSON sends a GET request to endpoint and decodes its JSON body into T.
// Using a type parameter lets the same HTTP and error-handling logic decode
// each Open-Meteo response type.
func fetchJSON[T any](ctx context.Context, endpoint string) (T, error) {
	// Return the zero value of T along with any error if the operation fails.
	var result T

	// Attach ctx to the request so cancellation and deadlines from the MCP
	// client propagate to the outbound HTTP request.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return result, fmt.Errorf("create API request: %w", err)
	}
	// Open-Meteo uses GET query parameters rather than a JSON request body, so
	// logging the URL captures the API request parameters.
	log.Printf("API request: %s %s", req.Method, req.URL)
	// Execute the request using the shared client (which enforces its timeout).
	resp, err := httpClient.Do(req)
	if err != nil {
		return result, fmt.Errorf("call API: %w", err)
	}
	// Always close the response body to release network resources.
	defer resp.Body.Close()

	// Read at most 4 KB from an error response so it cannot consume unbounded
	// memory; successful API responses are decoded and logged in full.
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 4096))
		if readErr != nil {
			return result, fmt.Errorf("API returned HTTP %d and its error response could not be read: %w", resp.StatusCode, readErr)
		}
		log.Printf("API response: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
		return result, fmt.Errorf("API returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, fmt.Errorf("read API response: %w", err)
	}
	log.Printf("API response: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	// Decode the successful JSON response directly into the caller's chosen T.
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&result); err != nil {
		return result, fmt.Errorf("decode API response: %w", err)
	}
	// Return the populated value and nil to indicate success.
	return result, nil
}

// main creates the MCP server, registers its weather tool, and serves MCP
// messages over standard input and output until the client disconnects.
func main() {
	// The standard logger writes to stderr by default. Keep that explicit for
	// stdio MCP so logs never mix with protocol messages on stdout.
	log.SetOutput(os.Stderr)
	if logPath := os.Getenv("WEATHER_MCP_LOG_FILE"); logPath != "" {
		logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			log.Fatalf("open log file %q: %v", logPath, err)
		}
		defer func() {
			if err := logFile.Close(); err != nil {
				log.Printf("close log file %q: %v", logPath, err)
			}
		}()
		log.SetOutput(io.MultiWriter(os.Stderr, logFile))
	}

	// Identify this server and its version to MCP clients.
	server := mcp.NewServer(&mcp.Implementation{Name: "open-meteo-weather", Version: "1.0.0"}, nil)

	// Register the tool's name and description, and bind it to the handler.
	// The SDK derives the tool's input schema from WeatherInput.
	mcp.AddTool(server, &mcp.Tool{
		Name: "get_current_weather",
		Description: "Get today's weather forecast for a city or place. Present the returned follow_up_question to the user. " +
			"If the user wants the seven-day forecast, call this tool again with the same location and include_seven_day_forecast=true.",
	}, getCurrentWeather)

	// StdioTransport uses stdin/stdout for MCP protocol messages. Do not print
	// ordinary logs to stdout, because they would corrupt that protocol stream.
	log.Printf("starting open-meteo-weather MCP server over stdio")
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		// Terminate with a non-zero status if the MCP server cannot run.
		log.Fatalf("MCP server stopped with an error: %v", err)
	}
	log.Printf("open-meteo-weather MCP server stopped")
}
