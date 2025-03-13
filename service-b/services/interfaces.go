package services

import (
	"context"
	"fullcycle-goexpert-desafio-temperature-for-cep/models"
)

type CEPService interface {
	GetCityByCEP(ctx context.Context, cep string) (string, error)
}

type WeatherService interface {
	GetTemperature(ctx context.Context, city string) (*models.Temperature, error)
}
