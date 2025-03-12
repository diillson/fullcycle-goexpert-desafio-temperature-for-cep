package handlers

import (
	"context"
	"encoding/json"
	"fullcycle-goexpert-desafio-temperature-for-cep/models"
	"fullcycle-goexpert-desafio-temperature-for-cep/services"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"io"
	"log"
	"net/http"
)

type WeatherHandler struct {
	cepService     services.CEPService
	weatherService services.WeatherService
	tracer         trace.Tracer
}

type CepRequest struct {
	Cep string `json:"cep"`
}

type WeatherResponse struct {
	City  string  `json:"city"`
	TempC float64 `json:"temp_C"`
	TempF float64 `json:"temp_F"`
	TempK float64 `json:"temp_K"`
}

func NewWeatherHandler(cep services.CEPService, weather services.WeatherService) *WeatherHandler {
	return &WeatherHandler{
		cepService:     cep,
		weatherService: weather,
		tracer:         otel.Tracer("weather-handler"),
	}
}

func (h *WeatherHandler) GetWeatherByCEP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := h.tracer.Start(ctx, "GetWeatherByCEP")
	defer span.End()

	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	cep := vars["cep"]

	log.Printf("Recebida requisição para CEP: %s", cep)
	span.SetAttributes(attribute.String("cep", cep))

	h.processWeatherRequest(ctx, w, cep)
}

func (h *WeatherHandler) GetWeatherByCEPPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := h.tracer.Start(ctx, "GetWeatherByCEPPost")
	defer span.End()

	w.Header().Set("Content-Type", "application/json")

	var req CepRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
		return
	}

	if err := json.Unmarshal(body, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request format"})
		return
	}

	log.Printf("Recebida requisição POST para CEP: %s", req.Cep)
	span.SetAttributes(attribute.String("cep", req.Cep))

	h.processWeatherRequest(ctx, w, req.Cep)
}

func (h *WeatherHandler) processWeatherRequest(ctx context.Context, w http.ResponseWriter, cep string) {
	ctx, span := h.tracer.Start(ctx, "processWeatherRequest")
	defer span.End()

	if len(cep) != 8 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid zipcode"})
		return
	}

	// Span para a busca de cidade por CEP
	var city string
	var err error
	func() {
		ctx, span := h.tracer.Start(ctx, "GetCityByCEP")
		defer span.End()

		city, err = h.cepService.GetCityByCEP(ctx, cep)
	}()

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

	// Span para a busca de temperatura
	var temp *models.Temperature
	func() {
		ctx, span := h.tracer.Start(ctx, "GetTemperature")
		defer span.End()

		temp, err = h.weatherService.GetTemperature(ctx, city)
	}()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to get weather data"})
		return
	}

	response := WeatherResponse{
		City:  city,
		TempC: temp.TempC,
		TempF: temp.TempF,
		TempK: temp.TempK,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
