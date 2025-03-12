package main

import (
	"context"
	"fullcycle-goexpert-desafio-temperature-for-cep/handlers"
	"fullcycle-goexpert-desafio-temperature-for-cep/services"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"log"
	"net/http"
	"os"
)

func initTracer() (*sdktrace.TracerProvider, error) {
	zipkinURL := os.Getenv("ZIPKIN_URL")
	if zipkinURL == "" {
		zipkinURL = "http://zipkin:9411/api/v2/spans"
	}

	exporter, err := zipkin.New(zipkinURL)
	if err != nil {
		return nil, err
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("service-b"),
		)),
	)

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tracerProvider, nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Inicializa o tracer
	tp, err := initTracer()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	cepService := &services.ViaCEPService{}
	weatherService := &services.WeatherAPIService{}
	handler := handlers.NewWeatherHandler(cepService, weatherService)

	r := mux.NewRouter()

	// Adiciona middleware do OpenTelemetry
	r.Use(otelmux.Middleware("service-b"))

	r.HandleFunc("/weather/{cep}", handler.GetWeatherByCEP).Methods("GET")
	r.HandleFunc("/weather", handler.GetWeatherByCEPPost).Methods("POST")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
