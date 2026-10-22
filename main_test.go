package main

import (
	"reflect"
	"testing"
)

func TestMapDailyForecast(t *testing.T) {
	daily := forecastDailyResponse{
		Time:                     []string{"2026-10-22", "2026-10-23", "2026-10-24", "2026-10-25", "2026-10-26", "2026-10-27", "2026-10-28"},
		TemperatureMin:           []float64{16, 17, 18, 19, 20, 21, 22},
		TemperatureMax:           []float64{24, 25, 26, 27, 28, 29, 30},
		PrecipitationProbability: []int{10, 20, 30, 40, 50, 60, 70},
		WeatherCode:              []int{1, 2, 3, 4, 5, 6, 7},
	}

	got, err := mapDailyForecast(daily, forecastDays)
	if err != nil {
		t.Fatalf("mapDailyForecast returned error: %v", err)
	}
	want := []DailyForecast{
		{Date: "2026-10-22", MinimumTemperatureCelsius: 16, MaximumTemperatureCelsius: 24, PrecipitationProbabilityPct: 10, WeatherCode: 1},
		{Date: "2026-10-23", MinimumTemperatureCelsius: 17, MaximumTemperatureCelsius: 25, PrecipitationProbabilityPct: 20, WeatherCode: 2},
		{Date: "2026-10-24", MinimumTemperatureCelsius: 18, MaximumTemperatureCelsius: 26, PrecipitationProbabilityPct: 30, WeatherCode: 3},
		{Date: "2026-10-25", MinimumTemperatureCelsius: 19, MaximumTemperatureCelsius: 27, PrecipitationProbabilityPct: 40, WeatherCode: 4},
		{Date: "2026-10-26", MinimumTemperatureCelsius: 20, MaximumTemperatureCelsius: 28, PrecipitationProbabilityPct: 50, WeatherCode: 5},
		{Date: "2026-10-27", MinimumTemperatureCelsius: 21, MaximumTemperatureCelsius: 29, PrecipitationProbabilityPct: 60, WeatherCode: 6},
		{Date: "2026-10-28", MinimumTemperatureCelsius: 22, MaximumTemperatureCelsius: 30, PrecipitationProbabilityPct: 70, WeatherCode: 7},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("mapDailyForecast() = %#v, want %#v", got, want)
	}
}

func TestMapDailyForecastRejectsIncompleteResponse(t *testing.T) {
	_, err := mapDailyForecast(forecastDailyResponse{
		Time: []string{"2026-10-22"},
	}, forecastDays)
	if err == nil {
		t.Fatal("mapDailyForecast accepted incomplete daily data")
	}
}

func TestMapTodayForecast(t *testing.T) {
	got, err := mapDailyForecast(forecastDailyResponse{
		Time:                     []string{"2026-10-22"},
		TemperatureMin:           []float64{16},
		TemperatureMax:           []float64{24},
		PrecipitationProbability: []int{70},
		WeatherCode:              []int{3},
	}, 1)
	if err != nil {
		t.Fatalf("mapDailyForecast returned error: %v", err)
	}
	want := []DailyForecast{{
		Date:                        "2026-10-22",
		MinimumTemperatureCelsius:   16,
		MaximumTemperatureCelsius:   24,
		PrecipitationProbabilityPct: 70,
		WeatherCode:                 3,
	}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("mapDailyForecast() = %#v, want %#v", got, want)
	}
}
