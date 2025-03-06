package services

import "fullcycle-goexpert-desafio-temperature-for-cep/models"

type CEPService interface {
	GetCityByCEP(cep string) (string, error)
}

type WeatherService interface {
	GetTemperature(city string) (*models.Temperature, error)
}
