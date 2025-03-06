package services

import (
	"encoding/json"
	"fmt"
	"fullcycle-goexpert-desafio-temperature-for-cep/models"
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

func (s *WeatherAPIService) GetTemperature(city string) (*models.Temperature, error) {
	apiKey := os.Getenv("WEATHER_API_KEY")
	if apiKey == "" {
		log.Printf("WEATHER_API_KEY não configurada")
		return nil, fmt.Errorf("weather API key not configured")
	}

	encodedCity := url.QueryEscape(city)
	url := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s", apiKey, encodedCity)

	resp, err := http.Get(url)
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
