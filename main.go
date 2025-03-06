package main

import (
	"fullcycle-goexpert-desafio-temperature-for-cep/handlers"
	"fullcycle-goexpert-desafio-temperature-for-cep/services"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cepService := &services.ViaCEPService{}
	weatherService := &services.WeatherAPIService{}
	handler := handlers.NewWeatherHandler(cepService, weatherService)

	r := mux.NewRouter()
	r.HandleFunc("/weather/{cep}", handler.GetWeatherByCEP).Methods("GET")

	port := ":8080"
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(port, r))
}
