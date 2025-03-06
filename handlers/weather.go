package handlers

import (
	"encoding/json"
	"fullcycle-goexpert-desafio-temperature-for-cep/services"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type WeatherHandler struct {
	cepService     services.CEPService
	weatherService services.WeatherService
}

func NewWeatherHandler(cep services.CEPService, weather services.WeatherService) *WeatherHandler {
	return &WeatherHandler{
		cepService:     cep,
		weatherService: weather,
	}
}

func (h *WeatherHandler) GetWeatherByCEP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	cep := vars["cep"]

	log.Printf("Recebida requisição para CEP: %s", cep)

	city, err := h.cepService.GetCityByCEP(cep)
	if err != nil {
		switch err.Error() {
		case "invalid zipcode":
			w.WriteHeader(http.StatusUnprocessableEntity)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid zipcode"})
		case "can not find zipcode":
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "can not find zipcode"})
		default:
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
		}
		return
	}

	temp, err := h.weatherService.GetTemperature(city)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to get weather data"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(temp)
}
