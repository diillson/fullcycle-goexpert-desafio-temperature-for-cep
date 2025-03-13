package services

import (
	"context"
	"encoding/json"
	"fmt"
	"fullcycle-goexpert-desafio-temperature-for-cep/models"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"log"
	"net/http"
	"net/url"
	"os"
)

type WeatherAPIService struct{}

type WeatherAPIResponse struct {
	Current struct {
		TempC float64 `json:"temp_c"`
	} `json:"current"`
	Error struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (s *WeatherAPIService) GetTemperature(ctx context.Context, city string) (*models.Temperature, error) {
	tracer := otel.Tracer("weather-api-service")
	ctx, span := tracer.Start(ctx, "WeatherAPI-GetTemperature")
	defer span.End()

	span.SetAttributes(attribute.String("city", city))

	apiKey := os.Getenv("WEATHER_API_KEY")
	if apiKey == "" {
		log.Printf("WEATHER_API_KEY não configurada")
		return nil, fmt.Errorf("weather API key not configured")
	}

	encodedCity := url.QueryEscape(city)
	reqUrl := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s", apiKey, encodedCity)

	span.SetAttributes(attribute.String("url", "https://api.weatherapi.com/v1/current.json"))

	req, err := http.NewRequestWithContext(ctx, "GET", reqUrl, nil)
	if err != nil {
		log.Printf("Erro ao criar requisição para WeatherAPI: %v", err)
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Erro ao fazer requisição para WeatherAPI: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	var weatherResp WeatherAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&weatherResp); err != nil {
		log.Printf("Erro ao decodificar resposta da WeatherAPI: %v", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Status code inválido da WeatherAPI: %d", resp.StatusCode)
		return nil, fmt.Errorf("weather API error: %s", weatherResp.Error.Message)
	}

	tempC := weatherResp.Current.TempC
	tempF := tempC*1.8 + 32
	tempK := tempC + 273.15

	span.SetAttributes(
		attribute.Float64("temp_c", tempC),
		attribute.Float64("temp_f", tempF),
		attribute.Float64("temp_k", tempK),
	)

	return &models.Temperature{
		TempC: round(tempC, 2),
		TempF: round(tempF, 2),
		TempK: round(tempK, 2),
	}, nil
}

func round(num float64, places int) float64 {
	factor := float64(1)
	for i := 0; i < places; i++ {
		factor *= 10
	}
	return float64(int(num*factor+0.5)) / factor
}
